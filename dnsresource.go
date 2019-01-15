package main

// DnsResourceType defines if resource is Zone or Record
type DnsResourceType int

const (
	Zone   DnsResourceType = 0
	Record DnsResourceType = 1
)

// DnsResource defines a resource
type DnsResource struct {
	Key  string
	Type DnsResourceType
}
