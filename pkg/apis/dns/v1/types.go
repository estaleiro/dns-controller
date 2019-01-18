package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSZone describes a DNSZone resource
type DNSZone struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the custom resource spec
	Spec DNSZoneSpec `json:"spec"`
}

// DNSZoneSpec is the spec for a DNSZone resource
type DNSZoneSpec struct {
	ZoneName string `json:"zoneName"`
	Refresh int `json:"refresh"`
	Retry int `json:"retry"`
	Expire int `json:"expire"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSZoneList is a list of DNSZone resources
type DNSZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DNSZone `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSRecord describes a DNSRecord resource
type DNSRecord struct {
        // TypeMeta is the metadata for the resource, like kind and apiversion
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        // Spec is the custom resource spec
        Spec DNSRecordSpec `json:"spec"`
}

// DNSRecordSpec is the spec for a DNSRecord resource
type DNSRecordSpec struct {
        ZoneName string `json:"zoneName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DNSRecordList is a list of Record resources
type DNSRecordList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata"`

        Items []DNSRecord `json:"items"`
}

