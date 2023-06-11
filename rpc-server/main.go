package main

import (
	"context"
	"log"
	"os"

	rpc "github.com/aaronsng/assignment_demo_2023/rpc-server/kitex_gen/rpc/imservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var (
	// Global variable here
	mdb = &MongoClient{}
)

func main() {
	// The URL locating the Mongo server is located through here.
	// When the HTTP server is first launched, Kubernetes / Docker would
	// load an environment variable specifying the URL of the RPC service
	ctx := context.Background()
	db_username := os.Getenv("USER_NAME")
	db_password := os.Getenv("USER_PWD")
	db_url := os.Getenv("DB_URL")

	err := mdb.InitClient(ctx, "mongodb://"+db_username+":"+db_password+"@"+db_url, "")
	if err != nil {
		log.Println(err.Error())
	}

	r, err := etcd.NewEtcdRegistry([]string{"etcd:2379"}) // r should not be reused.
	if err != nil {
		log.Fatal(err)
	}

	svr := rpc.NewServer(new(IMServiceImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "demo.rpc.server",
	}))

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
