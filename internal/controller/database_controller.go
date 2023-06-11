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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mysqlprovisionerv1beta1 "gitlab.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"gitlab.com/henrywhitaker3/mysql-provisioner/internal/db"
	"gitlab.com/henrywhitaker3/mysql-provisioner/internal/handlers"
)

var (
	fn string = "mysql-provisioner.henrywhitaker.com/propogate"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	h := handlers.NewDatabaseHandler(ctx, r.Client, req)
	return handlers.RunHandler(l, h)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlprovisionerv1beta1.Database{}).
		Complete(r)
}

func getDBForConnection(ctx context.Context, client client.Client, connRef mysqlprovisionerv1beta1.ConnectionRef) (*db.DB, error) {
	conn := &mysqlprovisionerv1beta1.Connection{}
	err := client.Get(ctx, types.NamespacedName{Namespace: connRef.Namespace, Name: connRef.Name}, conn)
	if err != nil {
		return nil, err
	}

	p, err := conn.Spec.PasswordSecretRef.GetPassword(ctx, client, conn.Namespace)
	if err != nil {
		return nil, err
	}
	db, err := db.NewDB(conn.Spec.User, p, conn.Spec.Host, conn.Spec.Port)
	if err != nil {
		return nil, err
	}

	return db, nil
}
