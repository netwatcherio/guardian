package agent

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Probe struct {
	Type      ProbeType          `json:"type"bson:"type"`
	ID        primitive.ObjectID `json:"id"bson:"_id"`
	Agent     primitive.ObjectID `json:"agent"bson:"agent"`
	Pending   time.Time          `json:"pending"bson:"pending"` // timestamp of when it was made pending / invalidate it after 10 minutes or so?
	CreatedAt time.Time          `bson:"createdAt"json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"json:"updatedAt"`
	Config    ProbeConfig        `bson:"config"json:"config"`
}

type ProbeConfig struct {
	Type     ProbeType `json:"type" bson:"type"`
	Target   string    `json:"target" bson:"target"`
	Duration int       `json:"duration" bson:"duration"`
	Count    int       `json:"count" bson:"count"`
	Interval int       `json:"interval" bson:"interval"`
	Server   bool      `bson:"server"json:"server"`
}

type ProbeType string

const (
	ProbeType_RPERF       ProbeType = "RPERF"
	ProbeType_MTR         ProbeType = "MTR"
	ProbeType_PING        ProbeType = "PING"
	ProbeType_SPEEDTEST   ProbeType = "SPEEDTEST"
	ProbeType_NETWORKINFO ProbeType = "NETINFO"
)
