package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ErrorFormat struct {
	ObjectID primitive.ObjectID `json:"objectID,omitempty"`
	Message  string             `json:"message,omitempty"`
	Error    error              `json:"error,omitempty"`
	Function string             `json:"function,omitempty"`
	Level    logrus.Level       `json:"level,omitempty"`
	Package  string             `json:"package,omitempty"`
}

func (e ErrorFormat) String() (string, error) {
	marshal, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	return string(marshal), nil
}

func (e ErrorFormat) ToError() error {
	e.Print()
	return fmt.Errorf(e.String())
}

func (e ErrorFormat) Print() {
	switch e.Level.String() {
	case "warning":
		logrus.Warn(e.String())
	case "error":
		logrus.Error(e.String())
	default:
		logrus.Info(e.String())
	}
}
