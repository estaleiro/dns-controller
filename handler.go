package main

import (
	"os"
	"path"
	"text/template"

	v1 "github.com/estaleiro/dns-controller/pkg/apis/dns/v1"
	log "github.com/sirupsen/logrus"
)

// Handler interface contains the methods that are required
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// ZoneHandler is a implementation of Handler for Zone
type ZoneHandler struct {
	zoneDirectory string
}

// Init handles any handler initialization
func (t *ZoneHandler) Init() error {
	log.Info("ZoneHandler.Init")
	return nil
}

// ObjectCreated is called when an object is created
func (t *ZoneHandler) ObjectCreated(obj interface{}) {
	zone := obj.(*v1.DNSZone)

	zoneName := zone.GetObjectMeta().GetName()

	// namespace_object_zone
	fileName := zone.GetNamespace() + "_" + zoneName

	zoneFile := path.Clean(t.zoneDirectory + "/" + fileName)

	// check if zone file exists and exit
	if _, err := os.Stat(zoneFile); !os.IsNotExist(err) {
		log.Infof("zone file already exists: %v, recreating", zoneFile)
		err = os.Remove(zoneFile)
		if err != nil {
			log.Errorf("error deleting zone file: %v", err)
			return
		}
	}

	// then create a new empty file
	file, err := os.Create(zoneFile)
	if err != nil {
		log.Errorf("error creating zone file: %v", err)
		return
	}

	corednsTemplate, _ := template.ParseFiles("coredns.tmpl")

	err = corednsTemplate.Execute(file, zoneName)
	if err != nil {
		log.Errorf("error executing template: %v", err)
	}

	defer file.Close()

	log.Infof("zone file %s created", zoneFile)
}

// ObjectDeleted is called when an object is deleted
func (t *ZoneHandler) ObjectDeleted(obj interface{}) {
	zone := obj.(*v1.DNSZone)

	zoneName := zone.GetObjectMeta().GetName()

	// namespace_object_zone
	fileName := zone.GetNamespace() + "_" + zoneName

	zoneFile := path.Clean(t.zoneDirectory + "/" + fileName)

	// check if zone file exists and remove it
	if _, err := os.Stat(zoneFile); !os.IsNotExist(err) {
		err = os.Remove(zoneFile)
		if err != nil {
			log.Errorf("error deleting zone file: %v", err)
		}
	}

	log.Infof("zone %s deleted", zoneName)
}

// ObjectUpdated is called when an object is updated
func (t *ZoneHandler) ObjectUpdated(objOld, objNew interface{}) {
	zone := objOld.(*v1.DNSZone)

	zoneName := zone.GetObjectMeta().GetName()

	t.ObjectDeleted(objOld)

	t.ObjectCreated(objNew)

	log.Infof("zone %s updated", zoneName)
}

// RecordHandler is a implementation of Handler for Record
type RecordHandler struct {
	zoneDirectory string
}

// Init handles any handler initialization
func (t *RecordHandler) Init() error {
	log.Info("RecordHandler.Init")
	return nil
}

// ObjectCreated is called when an object is created
func (t *RecordHandler) ObjectCreated(obj interface{}) {
	log.Info("record added")
}

// ObjectDeleted is called when an object is deleted
func (t *RecordHandler) ObjectDeleted(obj interface{}) {
	log.Infof("record deleted")
}

// ObjectUpdated is called when an object is updated
func (t *RecordHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("record updated")
}
