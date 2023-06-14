package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client *mongo.Client
}

type MongoMessage struct {
	// ID        *primitive.ObjectID `bson:"_id,omitempty"`
	Sender   string `bson:"sender,omitempty"`
	Text     string `bson:"text,omitempty"`
	SendTime int64  `bson:"sendtime,omitempty"`
}

func (cli *MongoClient) createMongoClient(ctx context.Context, uri string) error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected Successfully")
	cli.client = client
	return nil
}

func (cli *MongoClient) saveMessage(ctx context.Context, chatID string, message *MongoMessage) error {
	coll := cli.client.Database("chats_DB").Collection(chatID)
	_, err := coll.InsertOne(ctx, bson.D{{"sender", message.Sender}, {"text", message.Text}, {"sendtime", message.SendTime}})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (cli *MongoClient) getChat(ctx context.Context, chatID string) ([]*rpc.Message, error) {
	coll := cli.client.Database("chats_DB").Collection(chatID)

	options := options.Find().SetProjection(bson.M{"_id": 0})

	cursor, err := coll.Find(ctx, bson.M{}, options)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	var returnArray []*rpc.Message
	for _, result := range results {
		bsonBytes, _ := bson.Marshal(result)
		newMessage := &rpc.Message{}

		err = bson.Unmarshal(bsonBytes, &newMessage)
		if err != nil {
			log.Printf("error unpacking messages")
		}
		newMessage.Chat = chatID
		returnArray = append(returnArray, newMessage)
	}
	return returnArray, nil
}
