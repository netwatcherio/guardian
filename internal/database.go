package internal

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseConnection struct {
	URI         string
	DB          string
	MongoDB     *mongo.Database
	MongoClient *mongo.Client
	Logger      *logrus.Logger
}

/*
func (d *DatabaseConnection) NewConnection() *DatabaseConnection {
	d.Connect()

	if d.MongoDB != nil && d.MongoClient != nil {
		return d
	}

	d.Logger.Fatalf("Failed to Connect to database: %s", d.DB)

	return nil
}*/

func (d *DatabaseConnection) Connect() {
	var err error
	session := options.Client().ApplyURI(d.URI)
	if err != nil {
		d.Logger.Fatal(err)
	}
	d.MongoClient, err = mongo.Connect(context.TODO(), session)
	if err != nil {
		d.Logger.Fatal(err)
	}
	d.MongoDB = d.MongoClient.Database(d.DB)
	d.Logger.Infof("Successfully connected to database: %s", d.DB)
}
