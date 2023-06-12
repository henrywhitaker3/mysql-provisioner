package handlers

import (
	"context"

	"github.com/go-logr/logr"
	mysqlprovisionerv1beta1 "github.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"github.com/henrywhitaker3/mysql-provisioner/internal/db"
	"github.com/henrywhitaker3/mysql-provisioner/internal/misc"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	propFn string = "mysql-provisioner.henrywhitaker.com/propogate"
)

type Handler interface {
	// Check the object exists
	Get() error
	// Code to execute when the resource is created/updated
	CreateOrUpdate() error
	// Code to execute when the resource is deleted
	Delete() error
	// Is the objects deletion timestamp zero?
	DeletionTimestampIsZero() bool
	// Get the objects finalizers
	GetFinalizers() []string
	// Remove a given finalizer from the object
	RemoveFinalizer(string) error
	// Update the resource with a successful status
	SuccessStatus() error
	// Update the resource with a failed status
	ErrorStatus(error) error
	// The finalizer the handler should look for
	LookAtFinalizer() string
}

func RunHandler(l logr.Logger, h Handler) (reconcile.Result, error) {
	if err := h.Get(); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("processing resource")

	if !h.DeletionTimestampIsZero() {
		l.Info("resource being deleted")
		fn := h.LookAtFinalizer()
		if fn != "" && misc.ContainsString(h.GetFinalizers(), fn) {
			l.Info("processing resource finalizer")

			if err := h.Delete(); err != nil {
				h.ErrorStatus(err)
				return ctrl.Result{}, err
			}

			err := h.RemoveFinalizer(fn)
			return ctrl.Result{}, err
		}
	}

	if err := h.CreateOrUpdate(); err != nil {
		h.ErrorStatus(err)
		return ctrl.Result{}, err
	}

	err := h.SuccessStatus()
	return ctrl.Result{}, err
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
