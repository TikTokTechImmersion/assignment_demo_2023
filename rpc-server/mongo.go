package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	cli *mongo.Client
}

type MongoChat struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	ChatRoom string             `bson:"chatroom,omitempty"`
	Message  string             `bson:"message,omitempty"`
	Sender   string             `bson:"sender,omitempty"`
	SendTime string             `bson:"sendtime,omitempty"`
}

func (c *MongoClient) InitClient(ctx context.Context, address string, password string) error {
	clientOpts := options.Client().ApplyURI(address)
	client, err := mongo.Connect(ctx, clientOpts)

	if err != nil {
		log.Fatal(err)
	}

	c.cli = client
	return nil
}

func (c *MongoClient) ObtainChat(chat string) string {
	// Ensure that the chat ID is a consistent form
	chat = strings.ToLower(chat)
	parties := strings.Split(chat, ":")
	var finalChatID string

	if parties[0] < parties[1] {
		finalChatID = parties[0] + ":" + parties[1]
	} else {
		finalChatID = parties[1] + ":" + parties[0]
	}

	return finalChatID
}

func (c *MongoClient) FormatMessage(message *rpc.Message) *MongoChat {
	// Some messages may not be saved correctly with an empty time.
	// Here I'll be adding the time, and peg it according to the server timing
	newMessage := &MongoChat{
		ChatRoom: c.ObtainChat(message.GetChat()),
		Message:  message.GetText(),
		Sender:   message.GetSender(),
		SendTime: strconv.FormatInt(time.Now().Unix(), 10),
	}
	return newMessage
}

func (c *MongoClient) SaveMessage(ctx context.Context, message *rpc.Message) *rpc.SendResponse {
	coll := c.cli.Database("imService").Collection("im")

	// Try to find roomID. If no roomID then screw it and create a new roomID
	newMessage := *c.FormatMessage(message)
	fmt.Println(newMessage.ChatRoom, newMessage.Message, newMessage.SendTime)

	if _, insertErr := coll.InsertOne(ctx, newMessage); insertErr != nil {
		resp := rpc.NewSendResponse()
		resp.Code, resp.Msg = 404, "Can't save new response"
		log.Fatal(insertErr)
	}

	resp := rpc.NewSendResponse()
	resp.Code, resp.Msg = 0, "Successfully updated chat room"
	return resp
}

func (c *MongoClient) GetRoomByID(ctx context.Context, req *rpc.PullRequest) *rpc.PullResponse {
	coll := c.cli.Database("imService").Collection("im")

	// Set the search parameters
	chatId := req.GetChat()
	chatId = c.ObtainChat(chatId)

	// Set the order
	var order int
	if req.GetReverse() {
		order = -1
	} else {
		order = 1
	}

	filter := bson.M{"chatroom": chatId}
	pipeline := []bson.M{
		{"$match": filter},
		{"$sort": bson.M{"sendtime": order}},
		{"$skip": req.GetCursor()},
		{"$limit": req.GetLimit() + 1},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		// If no ErrNoDocuments, insert the new message inside
		resp := rpc.NewPullResponse()

		if err == mongo.ErrNoDocuments {
			log.Fatal(err)
			resp.Code, resp.Msg = 404, "Chat room not found"
			return resp
		} else {
			log.Fatal(err)
			resp.Code, resp.Msg = 404, "Error finding messages"
			return resp
		}
	}

	// Extract from the cursor
	var chats []MongoChat
	if err = cursor.All(ctx, &chats); err != nil {
		panic(err)
	}

	// Check if there's more
	var hasMore bool
	var nextCursor int64
	var newLength int
	if len(chats) > int(req.GetLimit()) {
		hasMore = true
		nextCursor = int64(req.GetCursor()) + int64(req.GetLimit())
		newLength = int(req.GetLimit())
	} else {
		hasMore = false
		nextCursor = int64(req.GetCursor()) + int64(len(chats))
		newLength = len(chats)
	}

	// Store it in the Messages kitex definition
	messages := make([]*rpc.Message, newLength)
	for i, chat := range chats {
		if i >= newLength {
			break
		}
		newMessage := &rpc.Message{}
		newMessage.Chat = chat.ChatRoom
		newMessage.Text = chat.Message
		newMessage.Sender = chat.Sender
		sendTime, _ := strconv.Atoi(chat.SendTime)
		newMessage.SendTime = int64(sendTime)
		messages[i] = newMessage
	}

	resp := &rpc.PullResponse{
		Code:       0,
		Msg:        "Successfully retrieved",
		Messages:   messages,
		HasMore:    &hasMore,
		NextCursor: &nextCursor,
	}

	return resp
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
