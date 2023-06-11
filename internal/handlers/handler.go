package handlers

import (
	"github.com/go-logr/logr"
	"gitlab.com/henrywhitaker3/mysql-provisioner/internal/misc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	fn string = "mysql-provisioner.henrywhitaker.com/propogate"
)

type Handler interface {
	// Check the object exists
	Get() error
	// Code to execute when the resource is first created
	Create() error
	// Code to execute when the resource is updated
	Update() error
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
}

func RunHandler(l logr.Logger, h Handler) (reconcile.Result, error) {
	if err := h.Get(); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("processing resource")

	if !h.DeletionTimestampIsZero() {
		l.Info("resource being deleted")
		if misc.ContainsString(h.GetFinalizers(), fn) {
			l.Info("processing resource finalizer")

			if err := h.Delete(); err != nil {
				return ctrl.Result{}, err
			}

			err := h.RemoveFinalizer(fn)
			return ctrl.Result{}, err
		}
	}

	if err := h.Create(); err != nil {
		h.ErrorStatus(err)
		return ctrl.Result{}, err
	}

	err := h.SuccessStatus()
	return ctrl.Result{}, err

	// TODO: handling for differing between update/delete
}
