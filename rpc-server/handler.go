package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
	_ "github.com/lib/pq"
)

func connectDB() *sql.DB {
	// Connect to PostgreSQL database
	// Code from https://www.calhoun.io/connecting-to-a-postgresql-database-with-gos-database-sql-package/
	const (
		host     = "postgres"
		port     = 5432
		user     = "postgres"
		dbname   = "assignment_demo_2023"
		password = "blank"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, dbConnectErr := sql.Open("postgres", psqlInfo)
	if dbConnectErr != nil {
		panic(dbConnectErr)
	}

	dbConnectErr = db.Ping()
	if dbConnectErr != nil {
		panic(dbConnectErr)
	}

	return db
}

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	resp := rpc.NewSendResponse()

	// Code loosely based on https://www.calhoun.io/inserting-records-into-a-postgresql-database-with-gos-database-sql-package/
	db := connectDB()
	insertStatement := "INSERT INTO messages (chat, sender, message) VALUES ($1, $2, $3);"

	_, err := db.Exec(insertStatement, req.Message.Chat, req.Message.Sender, req.Message.Text)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	resp.Code, resp.Msg = 0, "Message sent successfully"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	resp := rpc.NewPullResponse()

	resp.Code, resp.Msg = areYouLucky()
	return resp, nil
}

func areYouLucky() (int32, string) {
	if rand.Int31n(2) == 1 {
		return 0, "success"
	} else {
		return 500, "oops"
	}
}
