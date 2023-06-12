package handlers

import (
	"context"
	"time"

	mysqlprovisionerv1beta1 "github.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"github.com/henrywhitaker3/mysql-provisioner/internal/db"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DatabaseHandler struct {
	ctx    context.Context
	client client.Client
	req    ctrl.Request
	obj    *mysqlprovisionerv1beta1.Database
}

func NewDatabaseHandler(ctx context.Context, client client.Client, req ctrl.Request) *DatabaseHandler {
	return &DatabaseHandler{
		ctx:    ctx,
		client: client,
		req:    req,
	}
}

func (h *DatabaseHandler) Get() error {
	d := &mysqlprovisionerv1beta1.Database{}
	if err := h.client.Get(h.ctx, h.req.NamespacedName, d); err != nil {
		return err
	}
	h.obj = d
	return nil
}

func (h *DatabaseHandler) CreateOrUpdate() error {
	db, err := h.getDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.CreateDB(h.ctx, h.obj.Spec.Name); err != nil {
		return err
	}

	return nil
}

func (h *DatabaseHandler) Delete() error {
	db, err := h.getDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	if h.obj.Status.Created {
		if err := db.DropDB(h.ctx, h.obj.Spec.Name); err != nil {
			return err
		}
	}

	return nil
}

func (h *DatabaseHandler) DeletionTimestampIsZero() bool {
	return h.obj.DeletionTimestamp.IsZero()
}

func (h *DatabaseHandler) GetFinalizers() []string {
	return h.obj.GetFinalizers()
}

func (h *DatabaseHandler) RemoveFinalizer(finalizer string) error {
	controllerutil.RemoveFinalizer(h.obj, finalizer)
	return h.client.Update(h.ctx, h.obj)
}

func (h *DatabaseHandler) SuccessStatus() error {
	h.obj.Status = mysqlprovisionerv1beta1.DatabaseStatus{
		Created: true,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	return h.client.SubResource("status").Update(h.ctx, h.obj)
}

func (h *DatabaseHandler) ErrorStatus(err error) error {
	h.obj.Status = mysqlprovisionerv1beta1.DatabaseStatus{
		Created: true,
		Error:   err.Error(),
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	h.client.SubResource("status").Update(h.ctx, h.obj)
	return err
}

func (h *DatabaseHandler) LookAtFinalizer() string {
	return propFn
}

func (h *DatabaseHandler) getDatabase() (*db.DB, error) {
	db, err := getDBForConnection(h.ctx, h.client, h.obj.Spec.ConnRef)
	if err != nil {
		return nil, err
	}
	return db, nil
}
