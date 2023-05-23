package models

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler : wrapper for mongo.Client
type Handler struct {
	client *mongo.Client
}

// TODO check return dbHandler or error
// NewDBHandler : returns a new Handler
func NewDBHandler() *Handler {
	// TODO get connection string from a config file/ envirement file
	clientOptions := options.Client().ApplyURI("mongodb://0.0.0.0:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal("ERROR CREAING DB HANDLER")
		log.Fatal(err)
		return nil
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &Handler{client}
}

// CloseConnection : closes connection with mongo db
func (db *Handler) CloseConnection() {
	err := db.client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
		return
	}
}
