package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	zoneclientset "github.com/estaleiro/dns-controller/pkg/client/clientset/versioned"
	zoneinformerv1 "github.com/estaleiro/dns-controller/pkg/client/informers/externalversions/zone/v1"
	listers "github.com/estaleiro/dns-controller/pkg/client/listers/zone/v1"

	recordclientset "github.com/estaleiro/dns-controller/pkg/client/clientset/versioned"
	recordinformerv1 "github.com/estaleiro/dns-controller/pkg/client/informers/externalversions/zone/v1"
)

// retrieve the Kubernetes cluster client from outside of the cluster
func getKubernetesClient() (kubernetes.Interface, zoneclientset.Interface, recordclientset.Interface) {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	zoneClient, err := zoneclientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	recordClient, err := recordclientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	log.Info("Successfully constructed k8s client")
	return client, zoneClient, recordClient
}

func main() {
	// get the Kubernetes client for connectivity
	client, zoneClient, recordClient := getKubernetesClient()

	// retrieve our custom resource informer which was generated from
	// the code generator and pass it the custom resource client, specifying
	// we should be looking through all namespaces for listing and watching
	zoneInformer := zoneinformerv1.NewZoneInformer(
		zoneClient,
		metav1.NamespaceAll,
		0,
		cache.Indexers{},
	)

	recordInformer := recordinformerv1.NewRecordInformer(
		recordClient,
		metav1.NamespaceAll,
		0,
		cache.Indexers{},
	)

	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	zoneDeletedIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

	zoneInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("Add zone: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				queue.Add(DnsResource{Key: key, Type: Zone})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("Update zone: %s", key)
			if err == nil {
				queue.Add(DnsResource{Key: key, Type: Zone})
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("Delete zone: %s", key)
			if err == nil {
				zoneDeletedIndexer.Add(obj)
				queue.Add(DnsResource{Key: key, Type: Zone})
			}
		},
	})

	recordDeletedIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

	recordInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("Add record: %s", key)
			if err == nil {
				queue.Add(DnsResource{Key: key, Type: Record})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("Update record: %s", key)
			if err == nil {
				queue.Add(DnsResource{Key: key, Type: Record})
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("Delete record: %s", key)
			if err == nil {
				recordDeletedIndexer.Add(obj)
				queue.Add(DnsResource{Key: key, Type: Record})
			}
		},
	})

	// Get zone directory
	var zoneDirectory string
	flagSet := flag.NewFlagSetWithEnvPrefix(os.Args[0], "COREDNS", 0)
	flagSet.StringVar(&zoneDirectory, "zone_dir", "/tmp/zones/", "coredns zones directory path")
	flagSet.Parse(os.Args[1:])
	log.Infof("zone_dir: %s", zoneDirectory)

	controller := Controller{
		logger:               log.NewEntry(log.New()),
		clientset:            client,
		zoneInformer:         zoneInformer,
		zoneLister:           listers.NewZoneLister(zoneInformer.GetIndexer()),
		recordInformer:       recordInformer,
		recordLister:         listers.NewRecordLister(recordInformer.GetIndexer()),
		queue:                queue,
		zoneHandler:          &ZoneHandler{zoneDirectory: zoneDirectory},
		recordHandler:        &RecordHandler{zoneDirectory: zoneDirectory},
		zoneDeletedIndexer:   zoneDeletedIndexer,
		recordDeletedIndexer: recordDeletedIndexer,
	}

	// use a channel to synchronize the finalization for a graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	// run the controller loop to process items
	go controller.Run(stopCh)

	// use a channel to handle OS signals to terminate and gracefully shut
	// down processing
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
