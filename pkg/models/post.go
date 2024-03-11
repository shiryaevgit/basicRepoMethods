package models

import "time"

type Post struct {
	ID        int       `json:"id" bson:"_id"`
	UserId    int       `json:"userId" bson:"userId"`
	Text      string    `json:"text" bson:"text"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
