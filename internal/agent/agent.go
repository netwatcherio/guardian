package agent

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Agent struct {
	ID          primitive.ObjectID `bson:"_id, omitempty"json:"id"`       // id
	Name        string             `bson:"name"json:"name"form:"name"`    // name of the agentprobe
	Site        primitive.ObjectID `bson:"site"json:"site"`               // _id of mongo object
	Pin         string             `bson:"pin"json:"pin"`                 // used for registration & authentication
	Initialized bool               `bson:"initialized"json:"initialized"` // will this be used or will we use the sessions/jwt tokens?
	Location    float64            `bson:"location"json:"location"`       // logical/physical location
	CreatedAt   time.Time          `bson:"createdAt"json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"json:"updatedAt"`
	// pin will be used for "auth" as the password, the ID will stay the same
}
