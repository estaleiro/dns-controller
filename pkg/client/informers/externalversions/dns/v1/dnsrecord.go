/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	dnsv1 "github.com/estaleiro/dns-controller/pkg/apis/dns/v1"
	versioned "github.com/estaleiro/dns-controller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/estaleiro/dns-controller/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/estaleiro/dns-controller/pkg/client/listers/dns/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// DNSRecordInformer provides access to a shared informer and lister for
// DNSRecords.
type DNSRecordInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.DNSRecordLister
}

type dNSRecordInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewDNSRecordInformer constructs a new informer for DNSRecord type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewDNSRecordInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredDNSRecordInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredDNSRecordInformer constructs a new informer for DNSRecord type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredDNSRecordInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EstaleiroV1().DNSRecords(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EstaleiroV1().DNSRecords(namespace).Watch(options)
			},
		},
		&dnsv1.DNSRecord{},
		resyncPeriod,
		indexers,
	)
}

func (f *dNSRecordInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredDNSRecordInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *dNSRecordInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&dnsv1.DNSRecord{}, f.defaultInformer)
}

func (f *dNSRecordInformer) Lister() v1.DNSRecordLister {
	return v1.NewDNSRecordLister(f.Informer().GetIndexer())
}
