package main

import (
	"log"

	rpc "github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc/imservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	r, svrConnectErr := etcd.NewEtcdRegistry([]string{"etcd:2379"}) // r should not be reused.
	if svrConnectErr != nil {
		log.Fatal(svrConnectErr)
	}

	svr := rpc.NewServer(new(IMServiceImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "demo.rpc.server",
	}))

	// Runs the server
	svrConnectErr = svr.Run()
	if svrConnectErr != nil {
		log.Println(svrConnectErr.Error())
	}
}
