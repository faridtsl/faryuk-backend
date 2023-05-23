package db

import (
	"context"
	"log"

	"FaRyuk/config"
	"FaRyuk/internal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertSharing : inserts a result sharing in database
func (db *Handler) InsertSharing(r *types.Sharing) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")

	_, err := collection.InsertOne(context.TODO(), r)
	if err != nil {
		return err
	}
	return nil
}

// GetSharings : returns all sharings
func (db *Handler) GetSharings() []types.Sharing {
	var results []types.Sharing
	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
		return make([]types.Sharing, 0)
	}

	for cur.Next(context.TODO()) {
		var elem types.Sharing
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
			return make([]types.Sharing, 0)
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
		return make([]types.Sharing, 0)
	}

	cur.Close(context.TODO())
	return results
}

// RemoveSharingByID : removes a sharing by its ID
func (db *Handler) RemoveSharingByID(id string) bool {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

// UpdateSharing : update a sharing
func (db *Handler) UpdateSharing(r *types.Sharing) bool {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	return err == nil
}

// GetSharingByID : returns sharing by ID
func (db *Handler) GetSharingByID(id string) (types.Sharing, error) {
	var result types.Sharing
	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetSharingsByUser : returns sharings that were given to a user
func (db *Handler) GetSharingsByUser(search string) ([]types.Sharing, error) {
	var results []types.Sharing

	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")
	filter := bson.M{"userId": search}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Sharing, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Sharing
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Sharing, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// GetCurrentSharingsByUser : returns pending sharings of a user
func (db *Handler) GetCurrentSharingsByUser(search string) ([]types.Sharing, error) {
	var results []types.Sharing

	collection := db.client.Database(config.Cfg.Database.Name).Collection("sharing")
	filter := bson.M{"userId": search}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Sharing, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Sharing
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem.State == "Pending" {
			results = append(results, elem)
		}
	}

	if err := cur.Err(); err != nil {
		return make([]types.Sharing, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}
