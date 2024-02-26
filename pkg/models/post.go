package models

import "time"

type Post struct {
	ID        int       `json:"id"`
	UserId    int       `json:"userId"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}
