package user

import (
	"FaRyuk/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collection_name = "comment"

type MongoRepository struct {
	*mongo.Collection
	config *config.AppConfig
}

func NewMongoRepository(config *config.AppConfig) *MongoRepository {
	return &MongoRepository{
		createMongoDBCLient(&config.Database),
		config,
	}
}

// CloseConnection : closes connection with mongo db
func (repo *MongoRepository) CloseUserDBConnection() {
	err := repo.Database().Client().Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
		return
	}
}

func createMongoDBCLient(dbConfig *config.Database) *mongo.Collection {
	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", dbConfig.Host, dbConfig.Port)))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return client.Database(dbConfig.DbName).Collection(collection_name)
}
