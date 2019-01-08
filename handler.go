package main

import (
	"os"
	"text/template"

	v1 "github.com/estaleiro/dns-controller/pkg/apis/zone/v1"
	log "github.com/sirupsen/logrus"
)

// Handler interface contains the methods that are required
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// CoreDNSHandler is a sample implementation of Handler
type CoreDNSHandler struct {
	zoneDirectory string
}

// Init handles any handler initialization
func (t *CoreDNSHandler) Init() error {
	log.Info("CoreDNSHandler.Init")
	return nil
}

// ObjectCreated is called when an object is created
func (t *CoreDNSHandler) ObjectCreated(obj interface{}) {
	zone := obj.(*v1.Zone)

	zoneFile := t.zoneDirectory + "/" + zone.Spec.ZoneName

	// check if zone file exists and remove it
	if _, err := os.Stat(zoneFile); !os.IsNotExist(err) {
		err = os.Remove(zoneFile)
		if err != nil {
			log.Errorf("error deleting zone file: %v", err)
		}
	}

	// then create a new empty file
	file, err := os.Create(zoneFile)
	if err != nil {
		log.Errorf("error creating zone file: %v", err)
	}

	corednsTemplate, _ := template.ParseFiles("coredns.tmpl")

	err = corednsTemplate.Execute(file, zone.Spec.ZoneName)
	if err != nil {
		log.Errorf("error executing template: %v", err)
	}

	defer file.Close()

	log.Infof("zone %s added", zone.Spec.ZoneName)
}

// ObjectDeleted is called when an object is deleted
func (t *CoreDNSHandler) ObjectDeleted(obj interface{}) {
	zone := obj.(*v1.Zone)
	log.Infof("zone %s deleted", zone.Spec.ZoneName)
}

// ObjectUpdated is called when an object is updated
func (t *CoreDNSHandler) ObjectUpdated(objOld, objNew interface{}) {
	zone := objOld.(*v1.Zone)
	log.Infof("zone %s updated", zone.Spec.ZoneName)
}
