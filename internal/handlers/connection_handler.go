package handlers

import (
	"context"
	"fmt"
	"time"

	mysqlprovisionerv1beta1 "github.com/henrywhitaker3/mysql-provisioner/api/v1beta1"
	"github.com/henrywhitaker3/mysql-provisioner/internal/db"
	"github.com/henrywhitaker3/mysql-provisioner/internal/misc"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ConnectionHandler struct {
	ctx        context.Context
	client     client.Client
	dbClient   rest.Interface
	userClient rest.Interface
	req        ctrl.Request
	obj        *mysqlprovisionerv1beta1.Connection
}

func NewConnectionHandler(ctx context.Context, client client.Client, req ctrl.Request, db rest.Interface, user rest.Interface) *ConnectionHandler {
	return &ConnectionHandler{
		ctx:        ctx,
		client:     client,
		req:        req,
		dbClient:   db,
		userClient: user,
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
	if !misc.ContainsString(h.GetFinalizers(), h.LookAtFinalizer()) {
		controllerutil.AddFinalizer(h.obj, h.LookAtFinalizer())
		if err := h.client.Update(h.ctx, h.obj); err != nil {
			return err
		}
	}

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
	resp := h.userClient.Get().Namespace("").Do(h.ctx)
	fmt.Println("here")
	fmt.Println(resp)
	fmt.Println("here")
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
	return connFn
}

func (h *ConnectionHandler) getDatabase() (*db.DB, error) {
	db, err := getDBForConnection(h.ctx, h.client, mysqlprovisionerv1beta1.ConnectionRef{Name: h.obj.Name, Namespace: h.obj.Namespace})
	if err != nil {
		return nil, err
	}
	return db, nil
}
