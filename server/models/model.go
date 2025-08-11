package models

import "time"

type User struct { 
	ID string
	Username string
	PasswordHash string
	CreatedAt time.Time
}

type Message struct { 
	ID string
	SenderID string
	RecipientID string
	Content string
	CreatedAt int64
}