package main

import (
	"fmt"
	"log"

	rpc "github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc/imservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"

	"database/sql"

	_ "github.com/lib/pq"
)

const (
	host     = "postgres"
	port     = 5432
	user     = "postgres"
	dbname   = "assignment_demo_2023"
	password = "blank"
)

func main() {
	r, svrConnectErr := etcd.NewEtcdRegistry([]string{"etcd:2379"}) // r should not be reused.
	if svrConnectErr != nil {
		log.Fatal(svrConnectErr)
	}

	svr := rpc.NewServer(new(IMServiceImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "demo.rpc.server",
	}))

	// Connect to PostgreSQL database
	// Code from https://www.calhoun.io/connecting-to-a-postgresql-database-with-gos-database-sql-package/
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, dbConnectErr := sql.Open("postgres", psqlInfo)
	if dbConnectErr != nil {
		panic(dbConnectErr)
	}
	defer db.Close()

	dbConnectErr = db.Ping()
	if dbConnectErr != nil {
		panic(dbConnectErr)
	}

	// Runs the server
	svrConnectErr = svr.Run()
	if svrConnectErr != nil {
		log.Println(svrConnectErr.Error())
	}
}
