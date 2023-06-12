package handlers

import (
	"context"
	"time"

	mysqlprovisionerv1beta1 "github.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"github.com/henrywhitaker3/mysql-provisioner/internal/db"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ConnectionHandler struct {
	ctx        context.Context
	client     client.Client
	restClient dynamic.Interface
	req        ctrl.Request
	obj        *mysqlprovisionerv1beta1.Connection
}

func NewConnectionHandler(ctx context.Context, client client.Client, req ctrl.Request, r dynamic.Interface) *ConnectionHandler {
	return &ConnectionHandler{
		ctx:        ctx,
		client:     client,
		req:        req,
		restClient: r,
	}
}

func (h *ConnectionHandler) Get() error {
	d := &mysqlprovisionerv1beta1.Connection{}
	if err := h.client.Get(h.ctx, h.req.NamespacedName, d); err != nil {
		return err
	}
	h.obj = d
	return nil
}

func (h *ConnectionHandler) CreateOrUpdate() error {
	db, err := h.getDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(h.ctx); err != nil {
		return err
	}

	return nil
}

func (h *ConnectionHandler) Delete() error {
	// If we reach here, it's got the propogation finalizer in there
	// so delete evrything related to the connection.
	objs := []client.Object{}
	if err := h.getSubResources(&objs); err != nil {
		return err
	}

	for _, obj := range objs {
		if err := h.client.Delete(h.ctx, obj); err != nil {
			return err
		}
	}

	return nil
}

func (h *ConnectionHandler) DeletionTimestampIsZero() bool {
	return h.obj.DeletionTimestamp.IsZero()
}

func (h *ConnectionHandler) GetFinalizers() []string {
	return h.obj.GetFinalizers()
}

func (h *ConnectionHandler) RemoveFinalizer(finalizer string) error {
	controllerutil.RemoveFinalizer(h.obj, finalizer)
	return h.client.Update(h.ctx, h.obj)
}

func (h *ConnectionHandler) SuccessStatus() error {
	h.obj.Status = mysqlprovisionerv1beta1.ConnectionStatus{
		Status: true,
		Time:   time.Now().UTC().Format(time.RFC3339),
	}
	return h.client.SubResource("status").Update(h.ctx, h.obj)
}

func (h *ConnectionHandler) ErrorStatus(err error) error {
	h.obj.Status = mysqlprovisionerv1beta1.ConnectionStatus{
		Status: false,
		Error:  err.Error(),
		Time:   time.Now().UTC().Format(time.RFC3339),
	}
	h.client.SubResource("status").Update(h.ctx, h.obj)
	return err
}

func (h *ConnectionHandler) LookAtFinalizer() string {
	return propFn
}

// Get all resources that reference this connection
func (h *ConnectionHandler) getSubResources(objs *[]client.Object) error {
	// TODO: implement it
	return nil
}

func (h *ConnectionHandler) getDatabase() (*db.DB, error) {
	db, err := getDBForConnection(h.ctx, h.client, mysqlprovisionerv1beta1.ConnectionRef{Name: h.obj.Name, Namespace: h.obj.Namespace})
	if err != nil {
		return nil, err
	}
	return db, nil
}
