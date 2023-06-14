package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	timestamp := time.Now().Unix()
	chatID, err := formatChat(req.Message.GetChat())
	if err != nil {
		return nil, err
	}

	err = validateMessage(chatID, req.Message)
	if err != nil {
		return nil, err
	}

	newMessage := &MongoMessage{Sender: req.Message.GetSender(), Text: req.Message.GetText(), SendTime: timestamp}
	resp := rpc.NewSendResponse()
	err = db_client.saveMessage(ctx, chatID, newMessage)
	if err != nil {
		resp.Code, resp.Msg = 500, "unable to save messages"
		return resp, err
	}

	resp.Code, resp.Msg = 0, "message saved sucessfully"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	chatID, err := formatChat(req.GetChat())
	if err != nil {
		return nil, err
	}
	allMessages, err := db_client.getChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	limit := int64(req.GetLimit())
	if limit == 0 {
		limit = 10
	}

	reverse := req.GetReverse()

	start := req.GetCursor()
	end := start + limit

	if reverse {
		sort.Slice(allMessages, func(i, j int) bool {
			return allMessages[j].SendTime < allMessages[i].SendTime
		})
	} else {
		sort.Slice(allMessages, func(i, j int) bool {
			return allMessages[i].SendTime < allMessages[j].SendTime
		})
	}

	messagesLength := len(allMessages)
	if int(start) >= messagesLength {
		err = fmt.Errorf("invalid cursor")
		return nil, err
	}

	var nextCursor int64 = 0
	hasMore := false

	if int(end) >= messagesLength {
		allMessages = allMessages[start:messagesLength]
	} else {
		allMessages = allMessages[start:end]
		nextCursor = end
		hasMore = true
	}

	resp := rpc.NewPullResponse()
	resp.Messages = allMessages
	resp.NextCursor = &nextCursor
	resp.HasMore = &hasMore
	resp.Code, resp.Msg = 0, "success"

	return resp, nil
}

func formatChat(chatID string) (string, error) {
	chatID = strings.ToLower(chatID)
	users := strings.Split(chatID, ":")
	if len(users) != 2 {
		err := fmt.Errorf("incorrect chat id %s", chatID)
		return "", err
	}
	sort.Strings(users)
	return strings.Join(users, ":"), nil
}

func validateMessage(chatID string, message *rpc.Message) error {
	text := message.GetText()
	if text == "" {
		err := fmt.Errorf("empty text is not allowed")
		return err
	}
	sender := message.GetSender()
	if sender == "" {
		err := fmt.Errorf("empty sender is not allowed")
		return err
	}
	if !strings.Contains(chatID, strings.ToLower(sender)) {
		err := fmt.Errorf("sender does not exist in chat")
		return err
	}
	return nil
}
