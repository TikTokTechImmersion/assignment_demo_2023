package main

import (
	"context"
	"errors"
	"testing"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/stretchr/testify/assert"
)

func TestIMServiceImpl_Send(t *testing.T) {
	type args struct {
		ctx context.Context
		req *rpc.SendRequest
	}

	sendTest := &rpc.Message{
		Chat: "test1:test2"
		Text: "testing text"
		Sender: "test1"
		SendTime: strconv.FormatInt(time.Now().Unix(), 10)
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Empty request sent",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{},
			},
			wantErr: errors.New("Empty request sent"),
		},
		{
			name: "Normal send request",
			args: args{
				ctx: context.Background(),
				req: sendTest,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IMServiceImpl{}
			got, err := s.Send(tt.args.ctx, tt.args.req)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.NotNil(t, got)
		})
	}
}
