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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mysqlprovisionerv1beta1 "github.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"github.com/henrywhitaker3/mysql-provisioner/internal/handlers"
)

// ConnectionReconciler reconciles a Connection object
type ConnectionReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	DbClient   rest.Interface
	UserClient rest.Interface
}

//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=connections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=connections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=connections/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	h := handlers.NewConnectionHandler(ctx, r.Client, req, r.DbClient, r.UserClient)
	return handlers.RunHandler(l, h)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlprovisionerv1beta1.Connection{}).
		Complete(r)
}
