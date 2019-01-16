package main

// DNSResourceType defines if resource is Zone or Record
type DNSResourceType int

// Defines DNSResource types
const (
	Zone   DNSResourceType = 0
	Record DNSResourceType = 1
)

// DNSResource defines a resource
type DNSResource struct {
	Key  string
	Type DNSResourceType
}
