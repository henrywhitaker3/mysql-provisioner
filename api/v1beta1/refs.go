package v1beta1

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PasswordSecretRef struct {
	// The name of the secret
	Name string `json:"name"`
	// The key of the field containing the password
	Key string `json:"key"`
}

func (p *PasswordSecretRef) GetPassword(ctx context.Context, client client.Client, namespace string) (string, error) {
	s := &v1.Secret{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: namespace,
			Name:      p.Name,
		},
		s,
	)
	if err != nil {
		return "", err
	}

	pw, ok := s.Data[p.Key]
	if !ok {
		return "", err
	}

	return string(pw), nil
}

type ConnectionRef struct {
	// The name of the connection resource
	Name string `json:"name"`
	// The namespace the connection resource is in
	Namespace string `json:"namespace"`
}
