package db

import (
	"context"
	"log"

	"FaRyuk/config"
	"FaRyuk/internal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertRunner : inserts runner in the database
func (db *Handler) InsertRunner(r *types.Runner) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("runner")
	_, err := collection.InsertOne(context.TODO(), r)
	return err
}

// GetRunners : gets all runners
func (db *Handler) GetRunners() ([]types.Runner, error) {
	var results []types.Runner
	collection := db.client.Database(config.Cfg.Database.Name).Collection("runner")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.Runner, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.Runner
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Runner, 0), err
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Runner, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// RemoveRunnerByID : removes runner by ID
func (db *Handler) RemoveRunnerByID(id string) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("runner")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	return err
}

// UpdateRunner : updates runner
func (db *Handler) UpdateRunner(r *types.Runner) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("runner")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	return err
}

// GetRunnerByID : retrieves runner by ID
func (db *Handler) GetRunnerByID(id string) (types.Runner, error) {
	var result types.Runner
	collection := db.client.Database(config.Cfg.Database.Name).Collection("runner")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetRunnersByUserID : search runners by user
func (db *Handler) GetRunnersByUserID(search string) ([]types.Runner, error) {
	var results []types.Runner

	collection := db.client.Database(config.Cfg.Database.Name).Collection("runner")
	filter := bson.M{"owner": search}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Runner, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Runner
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Runner, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}
