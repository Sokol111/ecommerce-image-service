package image

import (
	"time"
)

type entity struct {
	ID         string `bson:"_id"`
	Version    int
	Alt        string
	OwnerType  string `bson:"ownerType"`
	OwnerId    string `bson:"ownerId"`
	Role       string
	Key        string
	Mime       string
	Size       int64
	Status     string
	CreatedAt  time.Time `bson:"createdAt"`
	ModifiedAt time.Time `bson:"modifiedAt"`
}
