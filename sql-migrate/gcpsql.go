//go:build go1.3
// +build go1.3

package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	cloudsqlconnmysql "cloud.google.com/go/cloudsqlconn/mysql/mysql"
	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"

	migrate "github.com/rubenv/sql-migrate"
)

const driverName = "gcpsql"

func init() {
	sql.Register(driverName, &gcpSQLDriver{&mysql.MySQLDriver{}})

	dialects[driverName] = gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}
	migrate.MigrationDialects[driverName] = dialects[driverName]
}

type gcpSQLDriver struct {
	d *mysql.MySQLDriver
}

func (d *gcpSQLDriver) Open(name string) (driver.Conn, error) {
	dialer, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return nil, err
	}

	mysql.RegisterDialContext(driverName,
		func(ctx context.Context, addr string) (net.Conn, error) {
			conn, dialErr := dialer.Dial(ctx, addr, cloudsqlconn.WithPrivateIP())
			if dialErr != nil {
				return nil, dialErr
			}

			return cloudsqlconnmysql.LivenessCheckConn{Conn: conn}, nil
		})

	return d.d.Open(name)
}
