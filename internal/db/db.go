package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	conn *sql.DB
}

// Create a new db instance
func NewDB(user, password, host string, port int) (*DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", user, password, host, port))
	if err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func (d *DB) Ping(ctx context.Context) error {
	if err := d.conn.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (d *DB) CreateDB(ctx context.Context, name string) error {
	_, err := d.conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", name))
	return err
}

func (d *DB) DropDB(ctx context.Context, name string) error {
	_, err := d.conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s;", name))
	return err
}

func (d *DB) CreateUser(ctx context.Context, name, password, host string) error {
	_, err := d.conn.ExecContext(ctx, "CREATE USER ?@? identified by ?", name, host, password)
	return err
}

func (d *DB) DropUser(ctx context.Context, name, host string) error {
	_, err := d.conn.ExecContext(ctx, "DROP USER ?@?", name, host)
	return err
}

// Close the connection
func (d *DB) Close() error {
	return d.conn.Close()
}
