package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

type ErrorFormat struct {
	ObjectID primitive.ObjectID `json:"objectID,omitempty"`
	Message  string             `json:"message,omitempty"`
	Error    error              `json:"error,omitempty"`
	Function string             `json:"function,omitempty"`
	Level    logrus.Level       `json:"level,omitempty"`
	Package  string             `json:"package,omitempty"`
}

func (e ErrorFormat) String() string {
	marshal, err := json.Marshal(e)
	if err != nil {
		return ""
	}

	return string(marshal)
}

func (e ErrorFormat) ToError() error {
	e.Print()
	// todo send logs over to loki??!??
	return fmt.Errorf(e.String())
}

func (e ErrorFormat) Print() {
	// todo send over to loki as well??
	if os.Getenv("DEBUG") == "true" {
		switch e.Level.String() {
		case "warning":
			logrus.Warn(e.String())
		case "error":
			logrus.Error(e.String())
		default:
			logrus.Info(e.String())
		}
	}
}
