package db

import (
	"context"
	"log"

	"FaRyuk/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Handler : wrapper for mongo.Client
type Handler struct {
	client *mongo.Client
}

// NewDBHandler : returns a new Handler
func NewDBHandler() *Handler {
	clientOptions := options.Client().ApplyURI(config.Cfg.Database.URI)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
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
