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
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mysqlprovisionerv1beta1 "gitlab.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
)

// ConnectionReconciler reconciles a Connection object
type ConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	c := &mysqlprovisionerv1beta1.Connection{}
	if err := r.Get(ctx, req.NamespacedName, c); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the object is being deleted
	if !c.ObjectMeta.DeletionTimestamp.IsZero() {
		l.Info("mysql-provisioner.henrywhitaker.com/connection being deleted")
	}

	s := &v1.Secret{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Namespace: c.Namespace,
			Name:      c.Spec.PasswordSecretRef.Name,
		},
		s,
	)
	if err != nil {
		c.Status = mysqlprovisionerv1beta1.ConnectionStatus{
			Status: false,
			Error:  err.Error(),
		}
		err := r.Status().Update(ctx, c)
		return ctrl.Result{}, err
	}

	p, ok := s.Data[c.Spec.PasswordSecretRef.Key]
	if !ok {
		c.Status = mysqlprovisionerv1beta1.ConnectionStatus{
			Status: false,
			Error:  "secret key not found",
		}
		err := r.Status().Update(ctx, c)
		return ctrl.Result{}, err
	}

	if err := testConnection(ctx, c, string(p)); err != nil {
		c.Status = mysqlprovisionerv1beta1.ConnectionStatus{
			Status: false,
			Error:  err.Error(),
		}
		err := r.Status().Update(ctx, c)
		return ctrl.Result{}, err
	}

	c.Status = mysqlprovisionerv1beta1.ConnectionStatus{
		Status: true,
	}
	err = r.Status().Update(ctx, c)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlprovisionerv1beta1.Connection{}).
		Complete(r)
}

func testConnection(ctx context.Context, conn *mysqlprovisionerv1beta1.Connection, password string) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", conn.Spec.User, password, conn.Spec.Host, conn.Spec.Port))
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
