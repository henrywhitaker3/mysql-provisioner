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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mysqlprovisionerv1beta1 "gitlab.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"gitlab.com/henrywhitaker3/mysql-provisioner/internal/misc"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql-provisioner.henrywhitaker.com,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	u := &mysqlprovisionerv1beta1.User{}
	if err := r.Get(ctx, req.NamespacedName, u); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Processing user")

	db, err := getDBForConnection(ctx, r.Client, u.Spec.ConnRef)
	if err != nil {
		u.Status = mysqlprovisionerv1beta1.UserStatus{
			Created: false,
			Error:   err.Error(),
		}
		err := r.Status().Update(ctx, u)
		return ctrl.Result{}, err
	}
	defer db.Close()

	// Check if the object is being deleted
	if !u.ObjectMeta.DeletionTimestamp.IsZero() {
		if misc.ContainsString(u.GetFinalizers(), fn) {
			l.Info("propogating user deletion")
			if u.Status.Created {
				if err := db.DropUser(ctx, u.Spec.Name, u.Spec.Host); err != nil {
					u.Status = mysqlprovisionerv1beta1.UserStatus{
						Created: u.Status.Created,
						Error:   err.Error(),
					}
					err := r.Status().Update(ctx, u)
					return ctrl.Result{}, err
				}
			}

			controllerutil.RemoveFinalizer(u, fn)
			err := r.Update(ctx, u)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	pw, err := u.Spec.PasswordSecretRef.GetPassword(ctx, r.Client, u.Namespace)
	if err != nil {
		u.Status = mysqlprovisionerv1beta1.UserStatus{
			Created: false,
			Error:   err.Error(),
		}
		err := r.Status().Update(ctx, u)
		return ctrl.Result{}, err
	}

	if err := db.CreateUser(ctx, u.Spec.Name, pw, u.Spec.Host); err != nil {
		u.Status = mysqlprovisionerv1beta1.UserStatus{
			Created: false,
			Error:   err.Error(),
		}
		err := r.Status().Update(ctx, u)
		return ctrl.Result{}, err
	}

	u.Status = mysqlprovisionerv1beta1.UserStatus{
		Created: true,
	}
	err = r.Status().Update(ctx, u)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlprovisionerv1beta1.User{}).
		Complete(r)
}
