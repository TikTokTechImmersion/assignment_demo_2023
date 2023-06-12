package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

type dbConnectionParam struct {
	host     string
	port     int
	user     string
	dbname   string
	password string
}

var connectionParam = dbConnectionParam{
	host:     "postgres",
	port:     5432,
	user:     "postgres",
	dbname:   "assignment_demo_2023",
	password: "blank",
}

var db *sql.DB

// Check if running in go test: https://stackoverflow.com/questions/14249217/how-do-i-know-im-running-within-go-test
func init() {
	if strings.HasSuffix(os.Args[0], ".test") {
		db = nil
	} else {
		db = connectDB(&connectionParam)
	}
}

func connectDB(connectionParam *dbConnectionParam) *sql.DB {
	// Connect to PostgreSQL database
	// Code from https://www.calhoun.io/connecting-to-a-postgresql-database-with-gos-database-sql-package/
	host := connectionParam.host
	port := connectionParam.port
	user := connectionParam.user
	dbname := connectionParam.dbname
	password := connectionParam.password

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
