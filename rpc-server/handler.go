package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	_ "github.com/lib/pq"
	"github.com/relvacode/iso8601"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	return s.SendSpecifyingDatabase(ctx, req, db)
}

func (s *IMServiceImpl) SendSpecifyingDatabase(ctx context.Context, req *rpc.SendRequest, db *sql.DB) (*rpc.SendResponse, error) {
	resp := rpc.NewSendResponse()

	// Code loosely based on https://www.calhoun.io/inserting-records-into-a-postgresql-database-with-gos-database-sql-package/
	insertStatement := "INSERT INTO messages (chat, sender, message) VALUES ($1, $2, $3);"

	const invalidChatParam = "Chat parameter should be in the form <member1>:<member2>" +
		"and the sender should be either <member1> or <member2>"
	const invalidParamErrorCode = 1

	if strings.Count(req.Message.Chat, ":") != 1 {
		resp.Code = invalidParamErrorCode
		resp.Msg = invalidChatParam
		return resp, nil
	}

	user1, user2, _ := strings.Cut(req.Message.Chat, ":")

	if req.Message.Sender != user1 && req.Message.Sender != user2 {
		resp.Code = invalidParamErrorCode
		resp.Msg = invalidChatParam
		return resp, nil
	}

	_, err := db.Exec(insertStatement, req.Message.Chat, req.Message.Sender, req.Message.Text)
	if err != nil {
		panic(err)
	}

	resp.Code, resp.Msg = 0, "Message sent successfully"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	return s.PullSpecifyingDatabase(ctx, req, db)
}

func (s *IMServiceImpl) PullSpecifyingDatabase(ctx context.Context, req *rpc.PullRequest, db *sql.DB) (*rpc.PullResponse, error) {
	resp := rpc.NewPullResponse()

	// Code loosely based on https://www.calhoun.io/querying-for-multiple-records-with-gos-sql-package/
	// Splicing in Go: https://stackoverflow.com/questions/7933460/how-do-you-write-multiline-strings-in-go
	sender, receiver, _ := strings.Cut(req.Chat, ":")

	var correctedChatParam string
	if strings.Compare(sender, receiver) < 0 {
		correctedChatParam = sender + ":" + receiver
	} else {
		correctedChatParam = receiver + ":" + sender
	}

	var order string
	if *req.Reverse {
		order = "ASC"
	} else {
		order = "DESC"
	}

	// To select from index n (1-based onwards) would mean an offset of n-1.
	// Reference: https://stackoverflow.com/questions/16568/how-to-select-the-nth-row-in-a-sql-database-table
	rows, err := db.Query(fmt.Sprintf(`SELECT message_id, sender, message, message_send_time FROM messages WHERE chat=$1 
		ORDER BY message_send_time %s LIMIT $2 OFFSET $3`, order),
		correctedChatParam, req.Limit+1, req.Cursor)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var message_id int
		var sender string
		var message string
		var message_send_time string
		err = rows.Scan(&message_id, &sender, &message, &message_send_time)
		if err != nil {
			panic(err)
		}
		message_send_time_golang, errTimeParse := iso8601.ParseString(message_send_time)
		if errTimeParse != nil {
			log.Default().Println(errTimeParse)
		}
		message_send_time_unix_nano := message_send_time_golang.UnixNano()

		newMessage := &rpc.Message{
			Chat:     correctedChatParam,
			Text:     message,
			Sender:   sender,
			SendTime: message_send_time_unix_nano,
		}
		resp.Messages = append(resp.Messages, newMessage)
	}

	hasMore := len(resp.Messages) > int(req.Limit)

	// convert boolean to pointer of boolean
	// refer to https://stackoverflow.com/questions/28817992/how-to-set-bool-pointer-to-true-in-struct-literal
	resp.HasMore = &[]bool{hasMore}[0]
	if hasMore {
		// resp.Messages has one more row than required in this case
		resp.Messages = resp.Messages[:len(resp.Messages)-1]
		resp.NextCursor = &[]int64{(req.Cursor + int64(req.Limit))}[0]
	}

	// Errors can still happen while iterating through the rows
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	resp.Code = 0
	resp.Msg = "Messages retrieved successfully"
	return resp, nil
}
