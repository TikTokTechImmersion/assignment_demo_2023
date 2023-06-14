package main

import (
	"database/sql"
	"log"
	"fmt"
	_ "github.com/go-sql-driver/mysql"

)

type MySQLClient struct {
	db *sql.DB
}

func (c *MySQLClient) InitClient(ctx context.Context, dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	c.db = db
	return nil
}

func (c *MySQLClient) SaveMessage(ctx context.Context, roomID string, message *Message) error {
	stmt, err := c.db.PrepareContext(ctx, "INSERT INTO messages(room_id, text, timestamp) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, roomID, message.Text, message.Timestamp)
	if err != nil {
		return err
	}

	return nil
}

func (c *MySQLClient) GetMessagesByRoomID(ctx context.Context, roomID string, start, end int64, reverse bool) ([]*Message, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if reverse {
		rows, err = c.db.QueryContext(ctx, "SELECT * FROM messages WHERE room_id = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp DESC", roomID, start, end)
	} else {
		rows, err = c.db.QueryContext(ctx, "SELECT * FROM messages WHERE room_id = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp ASC", roomID, start, end)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.RoomID, &msg.Text, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}