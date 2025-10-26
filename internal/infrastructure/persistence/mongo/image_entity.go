package mongo

import (
	"time"
)

type imageEntity struct {
	ID         string    `bson:"_id"`
	Version    int       `bson:"version"`
	Alt        string    `bson:"alt"`
	OwnerType  string    `bson:"ownerType"`
	OwnerID    string    `bson:"ownerId"`
	Role       string    `bson:"role"`
	Key        string    `bson:"key"`
	Mime       string    `bson:"mime"`
	Size       int64     `bson:"size"`
	Status     string    `bson:"status"`
	CreatedAt  time.Time `bson:"createdAt"`
	ModifiedAt time.Time `bson:"modifiedAt"`
}
