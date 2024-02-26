package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
}
