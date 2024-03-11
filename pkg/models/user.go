package models

import "time"

type User struct {
	ID        int       `json:"id" bson:"_id"`
	Login     string    `json:"login" bson:"login"`
	FullName  string    `json:"fullName"bson:"fullName"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
