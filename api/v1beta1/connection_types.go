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
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type PasswordSecretRef struct {
	// The name of the secret
	Name string `json:"name"`
	// The key of the field containing the password
	Key string `json:"key"`
}

// ConnectionSpec defines the desired state of Connection
type ConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Hostname of the mysql instance
	Host string `json:"host"`
	// Port number the mysql instanceis listening on
	Port int `json:"port"`

	// The user to connect to the mysql instance with.
	User              string            `json:"user"`
	PasswordSecretRef PasswordSecretRef `json:"passwordSecretRef"`
}

// ConnectionStatus defines the observed state of Connection
type ConnectionStatus struct {
	// True if the connection is successful, false if not
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Connection is the Schema for the connections API
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec,omitempty"`
	Status ConnectionStatus `json:"status,omitempty"`
}

func (c *Connection) GetPassword(ctx context.Context, client client.Client) (string, error) {
	s := &v1.Secret{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: c.Namespace,
			Name:      c.Spec.PasswordSecretRef.Name,
		},
		s,
	)
	if err != nil {
		return "", err
	}

	p, ok := s.Data[c.Spec.PasswordSecretRef.Key]
	if !ok {
		return "", err
	}

	return string(p), nil
}

//+kubebuilder:object:root=true

// ConnectionList contains a list of Connection
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connection{}, &ConnectionList{})
}
