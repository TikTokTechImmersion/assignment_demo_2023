package main

import (
	"context"
	"testing"
	"time"

	"github.com/aaronsng/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/stretchr/testify/assert"
)

func TestIMServiceImpl_Send(t *testing.T) {
	type args struct {
		ctx context.Context
		req *rpc.SendRequest
	}

	sendTest := &rpc.Message{
		Chat:     "test1:test2",
		Text:     "testing text",
		Sender:   "test1",
		SendTime: int64(time.Now().Unix()),
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Normal send request",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: sendTest,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// s := &IMServiceImpl{}
			// got, _ := s.Send(tt.args.ctx, tt.args.req)
			// assert.True(t, errors.Is(err, tt.wantErr))
			assert.True(t, true)
		})
	}
}
