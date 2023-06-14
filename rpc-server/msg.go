package main

type Message struct {
	ID        int    `json:"id"`
	RoomID    string `json:"room_id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}