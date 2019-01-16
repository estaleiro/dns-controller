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
	var zoneDirectory string
	flagSet := flag.NewFlagSetWithEnvPrefix(os.Args[0], "COREDNS", 0)
	flagSet.StringVar(&zoneDirectory, "zone_dir", "/tmp/zones/", "coredns zones directory path")
	flagSet.Parse(os.Args[1:])
	log.Infof("zone_dir: %s", zoneDirectory)

	client, zoneClient, recordClient := getKubernetesClient()

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

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	recordDeletedIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

	zoneDeletedIndexer := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.Indexers{})

	recordInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("Add record: %s", key)
			if err == nil {
				queue.Add(DNSResource{Key: key, Type: Record})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("Update record: %s", key)
			if err == nil {
				queue.Add(DNSResource{Key: key, Type: Record})
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("Delete record: %s", key)
			if err == nil {
				recordDeletedIndexer.Add(obj)
				queue.Add(DNSResource{Key: key, Type: Record})
			}
		},
	})

	zoneInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("Add zone: %s", key)
			if err == nil {
				queue.Add(DNSResource{Key: key, Type: Zone})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("Update zone: %s", key)
			if err == nil {
				queue.Add(DNSResource{Key: key, Type: Zone})
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("Delete zone: %s", key)
			if err == nil {
				zoneDeletedIndexer.Add(obj)
				queue.Add(DNSResource{Key: key, Type: Zone})
			}
		},
	})

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

	stopCh := make(chan struct{})
	defer close(stopCh)

	go controller.Run(stopCh)

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
