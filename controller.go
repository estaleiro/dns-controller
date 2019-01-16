package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	v1 "github.com/estaleiro/dns-controller/pkg/apis/zone/v1"
	listers "github.com/estaleiro/dns-controller/pkg/client/listers/zone/v1"
	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller defines all we need to run controller
type Controller struct {
	logger               *log.Entry
	clientset            kubernetes.Interface
	queue                workqueue.RateLimitingInterface
	zoneInformer         cache.SharedIndexInformer
	zoneLister           listers.ZoneLister
	recordInformer       cache.SharedIndexInformer
	recordLister         listers.RecordLister
	zoneHandler          Handler
	recordHandler        Handler
	zoneDeletedIndexer   cache.Indexer
	recordDeletedIndexer cache.Indexer
}

// Run starts controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	defer c.queue.ShutDown()

	c.logger.Info("Controller.Run: initiating")

	go c.zoneInformer.Run(stopCh)
	go c.recordInformer.Run(stopCh)

	// do the initial synchronization (one time) to populate resources
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}

	c.logger.Info("Controller.Run: cache sync complete")

	// run the runWorker method every second with a stop channel
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced check if informer had finished to sync
func (c *Controller) HasSynced() bool {
	return c.zoneInformer.HasSynced() && c.recordInformer.HasSynced()
}

// runWorker executes the loop to process new items added to the queue
func (c *Controller) runWorker() {
	log.Info("Controller.runWorker: starting")

	for c.processNextItem() {
		log.Info("Controller.runWorker: processing next item")
	}

	log.Info("Controller.runWorker: completed")
}

// processNextItem retrieves each queued item and takes the
// necessary handler action based off of if the item was
// created or deleted
func (c *Controller) processNextItem() bool {
	log.Info("Controller.processNextItem: start")

	key, quit := c.queue.Get()

	if quit {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)

		var dnsResource DNSResource
		var ok bool

		if dnsResource, ok = obj.(DNSResource); !ok {
			c.queue.Forget(obj)
			c.logger.Errorf("Controller.processNextItem: expected string in workqueue but got %#v", obj)
			return nil
		}

		if dnsResource.Type == Zone {
			if err := c.syncZoneHandler(dnsResource); err != nil {
				c.logger.Errorf("Controller.processNextItem: error syncing zone '%s': %s", dnsResource.Key, err.Error())
				return err
			}
		} else {
			if err := c.syncRecordHandler(dnsResource); err != nil {
				c.logger.Errorf("Controller.processNextItem: error syncing record '%s': %s", dnsResource.Key, err.Error())
				return err
			}
		}
		c.logger.Infof("Controller.processNextItem: successfully synced '%s'", dnsResource.Key)

		return nil
	}(key)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	// keep the worker loop running by returning true
	return true
}

func (c *Controller) syncZoneHandler(dnsResource DNSResource) error {
	key := dnsResource.Key

	namespace, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		c.queue.AddRateLimited(dnsResource)
		return fmt.Errorf("invalid resource key %s", key)
	}

	zoneItem, zoneExists, err := c.zoneInformer.GetIndexer().GetByKey(key)
	if err != nil {
		if c.queue.NumRequeues(dnsResource) < 5 {
			c.queue.AddRateLimited(dnsResource)
			return fmt.Errorf("failed processing item with key %s with error %v, retrying", key, err)
		}
		c.queue.Forget(dnsResource)
		return fmt.Errorf("failed processing item with key %s with error %v, no more retries", key, err)
	}

	if !zoneExists {
		zoneItemDeleted, zoneExistsDeleted, err := c.zoneDeletedIndexer.GetByKey(key)

		if err != nil || !zoneExistsDeleted {
			c.zoneDeletedIndexer.Delete(key)
			c.queue.Forget(dnsResource)
			return fmt.Errorf("failed processing item with key %s with error %v, no more retries", key, err)
		}

		c.logger.Infof("Controller.syncZoneHandler: object deleted detected: %s", key)
		c.zoneHandler.ObjectDeleted(zoneItemDeleted)
		c.zoneDeletedIndexer.Delete(key)
		c.queue.Forget(dnsResource)
	} else {
		zoneReceived := zoneItem.(*v1.Zone)
		zoneToCreate := zoneReceived

		// Get all the zones
		zones, err := c.zoneLister.Zones(namespace).List(labels.Everything())
		if err != nil {
			return fmt.Errorf("failed processing item with key %s with error %v, no more retries", key, err)
		}

		// check if exists multiple crd asking to create same ZoneName
		for _, zoneFound := range zones {
			if zoneFound.Spec.ZoneName == zoneReceived.Spec.ZoneName && zoneFound.GetObjectMeta().GetName() != zoneReceived.GetObjectMeta().GetName() {
				c.logger.Infof("Controller.syncZoneHandler: object %v found defined zoneName %v too", zoneFound.GetObjectMeta().GetName(), zoneToCreate.Spec.ZoneName)
				// if we found a older crd
				if zoneToCreate.GetCreationTimestamp().After(zoneFound.GetCreationTimestamp().Time) {
					// then we use zoneFound
					zoneToCreate = zoneFound
				}
			}
		}

		c.logger.Infof("Controller.syncZoneHandler: object created detected: %v", zoneToCreate.GetObjectMeta().GetName())
		c.zoneHandler.ObjectCreated(zoneToCreate)
		c.queue.Forget(dnsResource)
	}

	return nil
}

func (c *Controller) syncRecordHandler(dnsResource DNSResource) error {
	key := dnsResource.Key

	namespace, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		c.queue.AddRateLimited(dnsResource)
		return fmt.Errorf("invalid resource key %s", key)
	}

	recordItem, recordExists, err := c.recordInformer.GetIndexer().GetByKey(key)
	if err != nil {
		if c.queue.NumRequeues(dnsResource) < 5 {
			c.queue.AddRateLimited(dnsResource)
			return fmt.Errorf("failed processing item with key %s with error %v, retrying", key, err)
		} else {
			c.queue.Forget(dnsResource)
			return fmt.Errorf("failed processing item with key %s with error %v, no more retries", key, err)
		}
	}

	if !recordExists {
		recordItemDeleted, recordExistsDeleted, err := c.recordDeletedIndexer.GetByKey(key)

		if err != nil || !recordExistsDeleted {
			c.recordDeletedIndexer.Delete(key)
			c.queue.Forget(dnsResource)
			return fmt.Errorf("failed processing item with key %s with error %v, no more retries", key, err)
		}

		c.logger.Infof("Controller.syncRecordHandler: object deleted detected: %s", key)
		c.recordHandler.ObjectDeleted(recordItemDeleted)
		c.recordDeletedIndexer.Delete(key)
		c.queue.Forget(dnsResource)
	} else {
		recordToCreate := recordItem.(*v1.Record)

		records, err := c.recordLister.Records(namespace).List(labels.Everything())
		if err != nil {
			return fmt.Errorf("failed processing item with key %s with error %v, no more retries", key, err)
		}

		// check if exists multiple crd asking to create same Record
		/*for _, recordFound := range records {

		}*/
		c.logger.Infof("Controller.syncRecordHandler: records found %v", records)

		c.logger.Infof("Controller.syncRecordHandler: object created detected: %v", recordToCreate)
		//c.recordHandler.ObjectCreated(recordToCreate)
		c.queue.Forget(dnsResource)
	}

	return nil
}
