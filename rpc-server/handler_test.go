package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
)

var testDBParam = dbConnectionParam{
	host:     "localhost",
	port:     5432,
	user:     "postgres",
	dbname:   "assignment_demo_2023_test",
	password: "blank",
}

var testDB *sql.DB

// Code to set up and clean up based on https://github.com/ory/dockertest and
// https://github.com/bignerdranch/BNR-Blog-Dockertest/blob/main/storage/postgres/postgres_adapter_test.go
func Setup() *dockertest.Resource {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}

	err = pool.Client.Ping()
	if err != nil {
		panic(err)
	}

	resource, err := pool.Run("postgres", "latest", []string{
		fmt.Sprintf("POSTGRES_PASSWORD=%s", testDBParam.password),
		fmt.Sprintf("POSTGRES_DB=%s", testDBParam.dbname),
	})
	if err != nil {
		panic(err)
	}

	// connectDB may panic if there is a database connection error
	// https://stackoverflow.com/questions/33167282/how-to-return-a-value-in-a-go-function-that-panics
	if err := pool.Retry(func() (errConnect error) {
		testDB = connectDB(&testDBParam)

		// Delete table if table already exists to ensure clean state of table
		testDB.Exec("DROP TABLE IF EXISTS messages;")

		// Execute SQL queries in a file in golang
		// https://stackoverflow.com/questions/38998267/how-to-execute-a-sql-file
		file, err := os.ReadFile("../sql/create_tables.sql")
		if err != nil {
			panic(err)
		}

		requests := strings.Split(string(file), ";")
		for _, request := range requests {
			_, err := testDB.Exec(request)
			if err != nil {
				panic(err)
			}
		}

		defer func() {
			errConnect := recover()
			if errConnect != nil {
				log.Default().Println("recovered from panic: ", errConnect)
			}
		}()
		return nil
	}); err != nil {
		panic(err)
	}

	for index, message := range testMessages {
		user1, user2, _ := strings.Cut(message.Chat, ":")
		testChatReverseMessages[index] = rpc.Message{
			Chat:     user2 + ":" + user1,
			Sender:   message.Sender,
			Text:     message.Text,
			SendTime: message.SendTime,
		}
	}

	return resource
}

func Cleanup(resource *dockertest.Resource) {
	err := resource.Close()
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	resource := Setup()
	m.Run()
	Cleanup(resource)
}

func TestIMServiceImpl_Send(t *testing.T) {
	type args struct {
		ctx context.Context
		req *rpc.SendRequest
	}
	const successCode = 0
	const failedCode = 1
	tests := []struct {
		name     string
		args     args
		wantErr  error
		wantCode int32
		wantMsg  string
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: &rpc.Message{
						Chat:   "user1:user2",
						Text:   "Hello world!",
						Sender: "user1",
					},
				},
			},
			wantErr:  nil,
			wantCode: successCode,
			wantMsg:  "Message sent successfully",
		},
		{
			name: "success: second user sending",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: &rpc.Message{
						Chat:   "user1:user2",
						Text:   "Welcome!",
						Sender: "user2",
					},
				},
			},
			wantErr:  nil,
			wantCode: successCode,
			wantMsg:  "Message sent successfully",
		},
		{
			name: "sender should be one of the users in chat",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: &rpc.Message{
						Chat:   "user1:user2",
						Text:   "Hello world!",
						Sender: "user3",
					},
				},
			},
			wantErr:  nil,
			wantCode: failedCode,
			wantMsg: "Chat parameter should be in the form <member1>:<member2>" +
				"and the sender should be either <member1> or <member2>",
		},
		{
			name: "invalid chat parameter",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: &rpc.Message{
						Chat:   "user1user2",
						Text:   "Hello world!",
						Sender: "user1",
					},
				},
			},
			wantErr:  nil,
			wantCode: failedCode,
			wantMsg: "Chat parameter should be in the form <member1>:<member2>" +
				"and the sender should be either <member1> or <member2>",
		},
		{
			name: "allow self chats",
			args: args{
				ctx: context.Background(),
				req: &rpc.SendRequest{
					Message: &rpc.Message{
						Chat:   "user1:user1",
						Text:   "Self chats should be allowed",
						Sender: "user1",
					},
				},
			},
			wantErr:  nil,
			wantCode: successCode,
			wantMsg:  "Message sent successfully",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IMServiceImpl{}
			got, err := s.SendSpecifyingDatabase(tt.args.ctx, tt.args.req, testDB)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantMsg, got.Msg)
		})
	}
}

const golangReferenceTimeString = "2006-01-02T15:04:05Z07:00"

// Return only first value in two-valued argument
// Reference: https://stackoverflow.com/questions/53018490/return-only-the-first-result-of-a-multiple-return-values-in-golang
func First[T, U any](firstVal T, _ U) T {
	return firstVal
}

func Second[T, U any](_ T, secondVal U) U {
	return secondVal
}

func ResetTestDB() {
	testDB.Exec("TRUNCATE TABLE messages")
}

func GetUnixNanoTime(timeString string) int64 {
	return First(time.Parse(golangReferenceTimeString, timeString)).UnixNano()
}

/*
In the database, Chat is guaranteed to be in the form <member1>:<member2> where member1 is
lexicographically smaller than member2
*/
var testMessages = []rpc.Message{
	{
		Chat:     "a:b",
		Text:     "hello!",
		Sender:   "a",
		SendTime: GetUnixNanoTime("2023-06-13T18:47:00Z"),
	},
	{
		Chat:     "a:b",
		Text:     "welcome back!",
		Sender:   "b",
		SendTime: GetUnixNanoTime("2023-06-13T18:47:01Z"),
	},
	{
		Chat:     "a:b",
		Text:     "How are you!",
		Sender:   "a",
		SendTime: GetUnixNanoTime("2023-06-13T18:47:02Z"),
	},
	{
		Chat:     "a:userC",
		Text:     "What's your name!",
		Sender:   "userC",
		SendTime: GetUnixNanoTime("2023-06-13T18:47:02Z"),
	},
	{
		Chat:     "userC:userD",
		Text:     "Where are we going out for dinner?",
		Sender:   "userD",
		SendTime: GetUnixNanoTime("2023-06-13T18:47:03Z"),
	},
}

var testChatReverseMessages = [20]rpc.Message{}

func InitPull() {
	ResetTestDB()

	for _, message := range testMessages {
		testDB.Exec("INSERT INTO messages (chat, sender, message, message_send_time) VALUES ($1, $2, $3, $4)",
			message.Chat, message.Sender, message.Text, time.Unix(0, message.SendTime).Format(golangReferenceTimeString))
	}
}

func TestIMServiceImpl_Pull(t *testing.T) {
	type args struct {
		ctx context.Context
		req *rpc.PullRequest
	}

	InitPull()

	const successCode = 0
	tests := []struct {
		name            string
		args            args
		wantErr         error
		wantCode        int32
		wantMsg         string
		wantMessageList []*rpc.Message
		wantHasMore     bool
		wantNextCursor  int64
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  0,
					Limit:   10,
					Reverse: &[]bool{false}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    false,
			wantNextCursor: -1,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testMessages[2]}[0],
				&[]rpc.Message{testMessages[1]}[0],
				&[]rpc.Message{testMessages[0]}[0],
			},
		},
		{
			name: "chat b:a is the same as a:b",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "b:a",
					Cursor:  0,
					Limit:   10,
					Reverse: &[]bool{false}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    false,
			wantNextCursor: -1,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testChatReverseMessages[2]}[0],
				&[]rpc.Message{testChatReverseMessages[1]}[0],
				&[]rpc.Message{testChatReverseMessages[0]}[0],
			},
		},
		{
			name: "reverse",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  0,
					Limit:   10,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    false,
			wantNextCursor: -1,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testMessages[0]}[0],
				&[]rpc.Message{testMessages[1]}[0],
				&[]rpc.Message{testMessages[2]}[0],
			},
		},
		{
			name: "limit",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  0,
					Limit:   2,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    true,
			wantNextCursor: 2,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testMessages[0]}[0],
				&[]rpc.Message{testMessages[1]}[0],
			},
		},
		{
			name: "cursor",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  1,
					Limit:   10,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    false,
			wantNextCursor: -1,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testMessages[1]}[0],
				&[]rpc.Message{testMessages[2]}[0],
			},
		},
		{
			name: "limit and cursor, next cursor does not reach the end",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  1,
					Limit:   12,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    false,
			wantNextCursor: -1,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testMessages[1]}[0],
				&[]rpc.Message{testMessages[2]}[0],
			},
		},
		{
			name: "limit and cursor, next cursor reaches the end",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  1,
					Limit:   2,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:        nil,
			wantCode:       successCode,
			wantMsg:        "Messages retrieved successfully",
			wantHasMore:    false,
			wantNextCursor: -1,
			wantMessageList: []*rpc.Message{
				&[]rpc.Message{testMessages[1]}[0],
				&[]rpc.Message{testMessages[2]}[0],
			},
		},
		{
			name: "limit 0, chat existent",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  0,
					Limit:   0,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:         nil,
			wantCode:        successCode,
			wantMsg:         "Messages retrieved successfully",
			wantHasMore:     true,
			wantNextCursor:  0,
			wantMessageList: []*rpc.Message{},
		},
		{
			name: "limit 0, chat non-existent",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:c",
					Cursor:  0,
					Limit:   0,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:         nil,
			wantCode:        successCode,
			wantMsg:         "Messages retrieved successfully",
			wantHasMore:     false,
			wantNextCursor:  -1,
			wantMessageList: []*rpc.Message{},
		},
		{
			name: "cursor past number of chat messages",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  4,
					Limit:   2,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:         nil,
			wantCode:        successCode,
			wantMsg:         "Messages retrieved successfully",
			wantHasMore:     false,
			wantNextCursor:  -1,
			wantMessageList: []*rpc.Message{},
		},
		{
			name: "cursor past number of chat messages and limit 0",
			args: args{
				ctx: context.Background(),
				req: &rpc.PullRequest{
					Chat:    "a:b",
					Cursor:  3,
					Limit:   0,
					Reverse: &[]bool{true}[0],
				},
			},
			wantErr:         nil,
			wantCode:        successCode,
			wantMsg:         "Messages retrieved successfully",
			wantHasMore:     false,
			wantNextCursor:  -1,
			wantMessageList: []*rpc.Message{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &IMServiceImpl{}
			got, err := s.PullSpecifyingDatabase(tt.args.ctx, tt.args.req, testDB)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantMsg, got.Msg)
			assert.Equal(t, len(tt.wantMessageList), len(got.Messages))

			for index, message := range tt.wantMessageList {
				assert.Equal(t, message, got.Messages[index])
			}

			if tt.wantNextCursor < 0 {
				assert.Nil(t, got.NextCursor)
			} else {
				assert.Equal(t, tt.wantNextCursor, *got.NextCursor)
			}
		})
	}

	ResetTestDB()
}
