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

// Controller struct defines how a controller should encapsulate
// logging, client connectivity, informing (list and watching)
// queueing, and handling of resource changes
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

// Run is the main path of execution for the controller loop
func (c *Controller) Run(stopCh <-chan struct{}) {

	// handle a panic with logging and exiting
	defer utilruntime.HandleCrash()
	// ignore new items in the queue but when all goroutines
	// have completed existing items then shutdown
	defer c.queue.ShutDown()

	c.logger.Info("Controller.Run: initiating")

	// run the informer to start listing and watching resources
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

// HasSynced allows us to satisfy the Controller interface
// by wiring up the informer's HasSynced method to it
func (c *Controller) HasSynced() bool {
	return c.zoneInformer.HasSynced() && c.recordInformer.HasSynced()
}

// runWorker executes the loop to process new items added to the queue
func (c *Controller) runWorker() {
	log.Info("Controller.runWorker: starting")

	// invoke processNextItem to fetch and consume the next change
	// to a watched or listed resource
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

	// stop the worker loop from running as this indicates we
	// have sent a shutdown message that the queue has indicated
	// from the Get method
	if quit {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.queue.Done(obj)

		var dnsResource DnsResource
		var ok bool

		if dnsResource, ok = obj.(DnsResource); !ok {
			c.queue.Forget(obj)
			c.logger.Errorf("expected string in workqueue but got %#v", obj)
			return nil
		}

		if err := c.syncZoneHandler(dnsResource.Key); err != nil {
			c.queue.AddRateLimited(dnsResource)
			c.logger.Errorf("error syncing '%s': %s, requeuing", dnsResource.Key, err.Error())
			return err
		}

		c.logger.Infof("Successfully synced '%s'", dnsResource.Key)
		return nil
	}(key)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	/*
		if keyRaw.Type == Zone {
			// assert the string out of the key (format `namespace/name`)
			zoneKeyRaw := keyRaw.Key

			// take the string key and get the object out of the indexer
			//
			// item will contain the complex object for the resource and
			// exists is a bool that'll indicate whether or not the
			// resource was created (true) or deleted (false)
			//
			// if there is an error in getting the key from the index
			// then we want to retry this particular queue key a certain
			// number of times (5 here) before we forget the queue key
			// and throw an error
			zoneItem, zoneExists, err := c.zoneInformer.GetIndexer().GetByKey(zoneKeyRaw)
			if err != nil {
				if c.queue.NumRequeues(key) < 5 {
					c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", key, err)
					c.queue.AddRateLimited(key)
				} else {
					c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
					c.queue.Forget(key)
					utilruntime.HandleError(err)
				}
			}

			// if the item doesn't exist then it was deleted and we need to fire off the handler's
			// ObjectDeleted method. but if the object does exist that indicates that the object
			// was created (or updated) so run the ObjectCreated method
			//
			// after both instances, we want to forget the key from the queue, as this indicates
			// a code path of successful queue key processing
			if !zoneExists {
				zoneItemDeleted, zoneExistsDeleted, err := c.zoneDeletedIndexer.GetByKey(zoneKeyRaw)

				if err != nil || !zoneExistsDeleted {
					c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
					c.zoneDeletedIndexer.Delete(key)
					c.queue.Forget(key)
					utilruntime.HandleError(err)
				}

				c.logger.Infof("Controller.processNextItem: object deleted detected: %s", zoneKeyRaw)
				c.zoneHandler.ObjectDeleted(zoneItemDeleted)
				c.zoneDeletedIndexer.Delete(key)
				c.queue.Forget(key)
			} else {
				c.logger.Infof("Controller.processNextItem: object created detected: %s", zoneKeyRaw)
				c.zoneHandler.ObjectCreated(zoneItem)
				c.queue.Forget(key)
			}

		} else {
			// assert the string out of the key (format `namespace/name`)
			recordKeyRaw := keyRaw.Key

			// take the string key and get the object out of the indexer
			//
			// item will contain the complex object for the resource and
			// exists is a bool that'll indicate whether or not the
			// resource was created (true) or deleted (false)
			//
			// if there is an error in getting the key from the index
			// then we want to retry this particular queue key a certain
			// number of times (5 here) before we forget the queue key
			// and throw an error
			recordItem, recordExists, err := c.recordInformer.GetIndexer().GetByKey(recordKeyRaw)
			if err != nil {
				if c.queue.NumRequeues(key) < 5 {
					c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", key, err)
					c.queue.AddRateLimited(key)
				} else {
					c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
					c.queue.Forget(key)
					utilruntime.HandleError(err)
				}
			}

			// if the item doesn't exist then it was deleted and we need to fire off the handler's
			// ObjectDeleted method. but if the object does exist that indicates that the object
			// was created (or updated) so run the ObjectCreated method
			//
			// after both instances, we want to forget the key from the queue, as this indicates
			// a code path of successful queue key processing
			if !recordExists {
				recordItemDeleted, recordExistsDeleted, err := c.recordDeletedIndexer.GetByKey(recordKeyRaw)

				if err != nil || !recordExistsDeleted {
					c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
					c.recordDeletedIndexer.Delete(key)
					c.queue.Forget(key)
					utilruntime.HandleError(err)
				}

				c.logger.Infof("Controller.processNextItem: object deleted detected: %s", recordKeyRaw)
				c.recordHandler.ObjectDeleted(recordItemDeleted)
				c.recordDeletedIndexer.Delete(key)
				c.queue.Forget(key)
			} else {
				c.logger.Infof("Controller.processNextItem: object created detected: %s", recordKeyRaw)
				c.recordHandler.ObjectCreated(recordItem)
				c.queue.Forget(key)
			}
		}*/

	// keep the worker loop running by returning true
	return true
}

func (c *Controller) syncZoneHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Controller.syncZoneHandler: invalid resource key %s", key))
		return nil
	}

	zoneItem, zoneExists, err := c.zoneInformer.GetIndexer().GetByKey(key)
	if err != nil {
		if c.queue.NumRequeues(key) < 5 {
			c.logger.Errorf("Controller.syncZoneHandler: Failed processing item with key %s with error %v, retrying", key, err)
			c.queue.AddRateLimited(key)
		} else {
			c.logger.Errorf("Controller.syncZoneHandler: Failed processing item with key %s with error %v, no more retries", key, err)
			c.queue.Forget(key)
			return err
		}
	}

	if !zoneExists {
		zoneItemDeleted, zoneExistsDeleted, err := c.zoneDeletedIndexer.GetByKey(key)

		if err != nil || !zoneExistsDeleted {
			c.logger.Errorf("Controller.syncZoneHandler: Failed processing item with key %s with error %v, no more retries", key, err)
			c.zoneDeletedIndexer.Delete(key)
			c.queue.Forget(key)
			return err
		}

		c.logger.Infof("Controller.syncZoneHandler: object deleted detected: %s", key)
		c.zoneHandler.ObjectDeleted(zoneItemDeleted)
		c.zoneDeletedIndexer.Delete(key)
		c.queue.Forget(key)
	} else {
		// Get all the zones
		zones, err := c.zoneLister.Zones(namespace).List(labels.Everything())
		if err != nil {
			c.logger.Errorf("Controller.syncZoneHandler: Failed processing item with key %s with error %v, no more retries", key, err)
			return err
		}

		zoneToCreate := zoneItem.(*v1.Zone)

		// check if exists multiple crd asking to create same ZoneName
		for _, zoneFound := range zones {
			if zoneFound.Spec.ZoneName == zoneToCreate.Spec.ZoneName && zoneFound.GetObjectMeta().GetName() != zoneToCreate.GetObjectMeta().GetName() {
				c.logger.Infof("Controller.syncZoneHandler: object found %v and object received %v asked to create same zone %v",
					zoneFound.GetNamespace()+"/"+zoneFound.GetObjectMeta().GetName(), namespace+"/"+name, zoneToCreate.Spec.ZoneName)
				// if we found a older crd
				if zoneToCreate.GetCreationTimestamp().After(zoneFound.GetCreationTimestamp().Time) {
					// then we use zoneFound
					zoneToCreate = zoneFound
				}
			}
		}

		c.logger.Infof("Controller.syncZoneHandler: object created detected: %v", zoneToCreate)
		//c.zoneHandler.ObjectCreated(zoneToCreate)
		c.queue.Forget(key)
	}

	return nil
}
