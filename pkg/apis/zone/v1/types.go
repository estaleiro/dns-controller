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

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Record describes a Record resource
type Record struct {
        // TypeMeta is the metadata for the resource, like kind and apiversion
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        // Spec is the custom resource spec
        Spec RecordSpec `json:"spec"`
}

// RecordSpec is the spec for a Record resource
type RecordSpec struct {
        ZoneName string `json:"zoneName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RecordList is a list of Record resources
type RecordList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata"`

        Items []Record `json:"items"`
}

