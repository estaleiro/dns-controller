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

	corednsTemplate, _ := template.ParseFiles("coredns.tmpl")

	file, err := os.OpenFile("/tmp/corednsconf", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	err = corednsTemplate.Execute(file, zone.Spec.ZoneName)
	if err != nil {
		log.Errorf("error executing template: %v", err)
	}

	defer file.Close()

	log.Infof("zone %s added to coredns configuration file", zone.Spec.ZoneName)
}

// ObjectDeleted is called when an object is deleted
func (t *TestHandler) ObjectDeleted(obj interface{}) {
	log.Info("TestHandler.ObjectDeleted")
}

// ObjectUpdated is called when an object is updated
func (t *TestHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("TestHandler.ObjectUpdated")
}
