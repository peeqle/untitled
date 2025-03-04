package models

type Message struct {
	Content   string
	Timestamp int64
	UserId    uint
	ChatId    string
}
