package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Zone describes a Zone resource
type Zone struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the custom resource spec
	Spec ZoneSpec `json:"spec"`
}

// ZoneSpec is the spec for a Zone resource
type ZoneSpec struct {
	ZoneName string `json:"zoneName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZoneList is a list of Zone resources
type ZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Zone `json:"items"`
}
