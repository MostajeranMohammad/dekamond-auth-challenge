package entities

import "time"

type User struct {
	Id        uint32
	Phone     string
	CreatedAt time.Time
}
