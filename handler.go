package main

import (
	"io/ioutil"
	"os"

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

// TestHandler is a sample implementation of Handler
type TestHandler struct{}

// Init handles any handler initialization
func (t *TestHandler) Init() error {
	log.Info("TestHandler.Init")
	return nil
}

// ObjectCreated is called when an object is created
func (t *TestHandler) ObjectCreated(obj interface{}) {
	zone := obj.(*v1.Zone)

	zoneFile, err := os.OpenFile("/tmp/zones", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Infof("error opening file: %v", err)
		d1 := []byte(zone.Spec.ZoneName + "\n")
		err := ioutil.WriteFile("/tmp/zones", d1, 0644)
		log.Infof("creating file")
		if err != nil {
			log.Errorf("error creating file: %v", err)
		}
		return
	}

	defer zoneFile.Close()

	log.Info("appending file")
	if _, err = zoneFile.WriteString(zone.Spec.ZoneName + "\n"); err != nil {
		log.Errorf("error appending file: %v", err)
	}

	log.Info("TestHandler.ObjectCreated")
}

// ObjectDeleted is called when an object is deleted
func (t *TestHandler) ObjectDeleted(obj interface{}) {
	log.Info("TestHandler.ObjectDeleted")
}

// ObjectUpdated is called when an object is updated
func (t *TestHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("TestHandler.ObjectUpdated")
}
