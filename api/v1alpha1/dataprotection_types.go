package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataProtectionSpec defines the desired state of DataProtection
type DataProtectionSpec struct {
	// +kubebuilder:validation:Enum=backup;restore
	Type string `json:"type"`
	// +kubebuilder:validation:Required
	DatabaseCRName string  `json:"databaseName"`
	Restore        Restore `json:"restore"`
}

type Restore struct {
	// +kubebuilder:validation:Required
	SnapshotId string `json:"snapshotId"`
}

// DataProtectionStatus defines the observed state of DataProtection
type DataProtectionStatus struct {
	Status      string `json:"status"`
	Type        string `json:"type"`
	OperationId string `json:"operationId"`
	// backup or the restore time
	Time           time.Time `json:"time"`
	DatabaseCRName string    `json:"databaseName"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataProtection is the Schema for the dataprotections API
type DataProtection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataProtectionSpec   `json:"spec,omitempty"`
	Status DataProtectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataProtectionList contains a list of DataProtection
type DataProtectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataProtection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataProtection{}, &DataProtectionList{})
}
