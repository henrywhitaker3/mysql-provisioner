/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type UserPermission struct {
	Privleges string `json:"privileges"`
	On        string `json:"to"`
}

// UserSpec defines the desired state of User
type UserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The name of the user to create
	Name string `json:"name"`
	// The host to limit user access to
	Host string `json:"host"`

	ConnRef ConnectionRef `json:"connectionRef"`

	PasswordSecretRef PasswordSecretRef `json:"passwordSecretRef"`

	// The permissions to grant the user
	Grants []UserPermission `json:"grants"`
}

// UserStatus defines the observed state of User
type UserStatus struct {
	Created bool   `json:"created"`
	Error   string `json:"error"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// User is the Schema for the users API
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
