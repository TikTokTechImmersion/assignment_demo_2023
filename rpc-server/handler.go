package main

import (
	"context"

	"github.com/aaronsng/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	resp := mdb.SaveMessage(ctx, req.Message)
	// resp.Code, resp.Msg = s.sendMessage()
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	resp := mdb.GetRoomByID(ctx, req)
	// resp.Code, resp.Mspg = areYouLucky(req.String())
	return resp, nil
}
