package main

import (
	"context"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/http-server/kitex_gen/rpc"
	"github.com/TikTokTechImmersion/assignment_demo_2023/http-server/kitex_gen/rpc/imservice"
	"github.com/TikTokTechImmersion/assignment_demo_2023/http-server/proto_gen/api"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var cli imservice.Client

func main() {
	r, err := etcd.NewEtcdResolver([]string{"etcd:2379"})
	if err != nil {
		log.Fatal(err)
	}
	cli = imservice.MustNewClient("demo.rpc.server",
		client.WithResolver(r),
		client.WithRPCTimeout(1*time.Second),
		client.WithHostPorts("rpc-server:8888"),
	)

	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})

	h.POST("/api/send", sendMessage)
	h.GET("/api/pull", pullMessage)

	h.Spin()
}

func sendMessage(ctx context.Context, c *app.RequestContext) {
	var req api.SendRequest
	err := c.Bind(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, "Failed to parse request body: %v", err)
		return
	}

	sender := c.Query("sender")
	receiver := c.Query("receiver")

	if sender == "" || receiver == "" {
		c.String(consts.StatusBadRequest, "Sender name and receiver name cannot be empty")
		return
	}

	if strings.Count(sender, ":") >= 1 || strings.Count(receiver, ":") >= 1 {
		c.String(consts.StatusBadRequest, "Sender name and receiver name cannot contain the character :")
		return
	}

	if strings.Compare(sender, receiver) < 0 {
		req.Chat = sender + ":" + receiver
	} else {
		req.Chat = receiver + ":" + sender
	}

	req.Text = c.Query("text")
	req.Sender = sender

	resp, err := cli.Send(ctx, &rpc.SendRequest{
		Message: &rpc.Message{
			Chat:   req.Chat,
			Text:   req.Text,
			Sender: req.Sender,
		},
	})
	if err != nil {
		c.String(consts.StatusInternalServerError, err.Error())
	} else if resp.Code != 0 {
		c.String(consts.StatusInternalServerError, resp.Msg)
	} else {
		c.Status(consts.StatusOK)
	}
}

const (
	defaultCursor = 0
	defaultLimit  = 10
)

func pullMessage(ctx context.Context, c *app.RequestContext) {
	var req api.PullRequest
	err := c.Bind(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, "Failed to parse request body: %v", err)
		return
	}

	req.Chat = c.Query("chat")
	tempCursor := c.Query("cursor")
	tempLimit := c.Query("limit")
	tempReverse := strings.ToLower(c.Query("reverse"))

	if strings.Count(req.Chat, ":") != 1 {
		c.String(consts.StatusBadRequest, "Chat parameter should be in the form <member1>:<member2>, denoting a chat between two users")
		return
	}

	if tempCursor == "" {
		req.Cursor = defaultCursor
	} else {
		cursorInt, errConv := strconv.Atoi(tempCursor)
		if errConv != nil {
			c.String(consts.StatusBadRequest, "Cursor of %s is not an integer", tempCursor)
			return
		}
		req.Cursor = int64(cursorInt)
	}

	if req.Cursor < 0 {
		c.String(consts.StatusBadRequest, "Cursor cannot be negative")
		return
	}

	if tempLimit == "" {
		req.Limit = defaultLimit
	} else {
		limitInt, errConv := strconv.Atoi(tempLimit)
		if errConv != nil {
			c.String(consts.StatusBadRequest, "Limit of %s is not an integer", tempLimit)
			return
		}
		req.Limit = int32(limitInt)
	}

	if req.Limit < 0 {
		c.String(consts.StatusBadRequest, "Limit cannot be negative")
		return
	} else if req.Limit == math.MaxInt32 {
		c.String(consts.StatusBadRequest, "Max supported value of limit is %d", math.MaxInt32-1)
		return
	}

	if tempReverse == "true" {
		req.Reverse = true
	} else if tempReverse == "false" || tempReverse == "" {
		req.Reverse = false
	} else {
		c.String(consts.StatusInternalServerError, "Invalid reverse parameter, it should be either true or false")
		return
	}

	resp, err := cli.Pull(ctx, &rpc.PullRequest{
		Chat:    req.Chat,
		Cursor:  req.Cursor,
		Limit:   req.Limit,
		Reverse: &req.Reverse,
	})
	if err != nil {
		c.String(consts.StatusInternalServerError, err.Error())
		return
	} else if resp.Code != 0 {
		c.String(consts.StatusInternalServerError, resp.Msg)
		return
	}
	messages := make([]*api.Message, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		messages = append(messages, &api.Message{
			Chat:     msg.Chat,
			Text:     msg.Text,
			Sender:   msg.Sender,
			SendTime: msg.SendTime,
		})
	}
	c.JSON(consts.StatusOK, &api.PullResponse{
		Messages:   messages,
		HasMore:    resp.GetHasMore(),
		NextCursor: resp.GetNextCursor(),
	})
}
