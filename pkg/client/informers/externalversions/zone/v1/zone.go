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

	zonev1 "github.com/estaleiro/dns-controller/pkg/apis/zone/v1"
	versioned "github.com/estaleiro/dns-controller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/estaleiro/dns-controller/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/estaleiro/dns-controller/pkg/client/listers/zone/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ZoneInformer provides access to a shared informer and lister for
// Zones.
type ZoneInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.ZoneLister
}

type zoneInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewZoneInformer constructs a new informer for Zone type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewZoneInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredZoneInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredZoneInformer constructs a new informer for Zone type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredZoneInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DnscontrollerV1().Zones(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DnscontrollerV1().Zones(namespace).Watch(options)
			},
		},
		&zonev1.Zone{},
		resyncPeriod,
		indexers,
	)
}

func (f *zoneInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredZoneInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *zoneInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&zonev1.Zone{}, f.defaultInformer)
}

func (f *zoneInformer) Lister() v1.ZoneLister {
	return v1.NewZoneLister(f.Informer().GetIndexer())
}