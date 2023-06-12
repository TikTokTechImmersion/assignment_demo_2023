package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

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
		file, err := ioutil.ReadFile("../sql/create_tables.sql")
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
			assert.Equal(t, got.Code, tt.wantCode)
			assert.Equal(t, got.Msg, tt.wantMsg)
		})
	}
}
