package handlers

import (
	"context"
	"time"

	mysqlprovisionerv1beta1 "gitlab.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"gitlab.com/henrywhitaker3/mysql-provisioner/internal/db"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type UserHandler struct {
	ctx    context.Context
	client client.Client
	req    ctrl.Request
	obj    *mysqlprovisionerv1beta1.User
}

func NewUserHandler(ctx context.Context, client client.Client, req ctrl.Request) *UserHandler {
	return &UserHandler{
		ctx:    ctx,
		client: client,
		req:    req,
	}
}

func (h *UserHandler) Get() error {
	d := &mysqlprovisionerv1beta1.User{}
	if err := h.client.Get(h.ctx, h.req.NamespacedName, d); err != nil {
		return err
	}
	h.obj = d
	return nil
}

func (h *UserHandler) Create() error {
	db, err := h.getDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	pw, err := h.obj.Spec.PasswordSecretRef.GetPassword(h.ctx, h.client, h.obj.Namespace)
	if err != nil {
		return nil
	}

	if err := db.CreateUser(h.ctx, h.obj.Spec.Name, pw, h.obj.Spec.Host); err != nil {
		return err
	}

	// TODO: add grants

	return nil
}

func (h *UserHandler) Update() error {
	return nil
}

func (h *UserHandler) Delete() error {
	db, err := h.getDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.DropUser(h.ctx, h.obj.Spec.Name, h.obj.Spec.Host); err != nil {
		return err
	}

	return nil
}

func (h *UserHandler) DeletionTimestampIsZero() bool {
	return h.obj.DeletionTimestamp.IsZero()
}

func (h *UserHandler) GetFinalizers() []string {
	return h.obj.GetFinalizers()
}

func (h *UserHandler) RemoveFinalizer(finalizer string) error {
	controllerutil.RemoveFinalizer(h.obj, finalizer)
	return h.client.Update(h.ctx, h.obj)
}

func (h *UserHandler) SuccessStatus() error {
	h.obj.Status = mysqlprovisionerv1beta1.UserStatus{
		Created: true,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	return h.client.SubResource("status").Update(h.ctx, h.obj)
}

func (h *UserHandler) ErrorStatus(err error) error {
	h.obj.Status = mysqlprovisionerv1beta1.UserStatus{
		Created: true,
		Error:   err.Error(),
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	return h.client.SubResource("status").Update(h.ctx, h.obj)
}

func (h *UserHandler) getDatabase() (*db.DB, error) {
	db, err := getDBForConnection(h.ctx, h.client, h.obj.Spec.ConnRef)
	if err != nil {
		return nil, h.ErrorStatus(err)
	}
	return db, nil
}
