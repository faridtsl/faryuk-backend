package models

import (
	"context"
	"log"

	"FaRyuk/internal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertGroup : inserts group in the database
func (db *Handler) InsertGroup(r types.Group) error {
	collection := db.client.Database("faryuk").Collection("group")
	_, err := collection.InsertOne(context.TODO(), r)
	return err
}

// GetGroups : gets all groups
func (db *Handler) GetGroups() ([]types.Group, error) {
	var results []types.Group
	collection := db.client.Database("faryuk").Collection("group")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.Group, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.Group
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Group, 0), err
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Group, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// RemoveGroupByID : removes group by ID
func (db *Handler) RemoveGroupByID(id string) error {
	collection := db.client.Database("faryuk").Collection("group")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	return err
}

// UpdateGroup : updates group
func (db *Handler) UpdateGroup(r types.Group) error {
	collection := db.client.Database("faryuk").Collection("group")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	return err
}

// GetGroupByID : retrieves group by ID
func (db *Handler) GetGroupByID(id string) (types.Group, error) {
	var result types.Group
	collection := db.client.Database("faryuk").Collection("group")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetGroupsByName : search groups by exact name
func (db *Handler) GetGroupsByName(search string) (types.Group, error) {
	var result types.Group
	result.ID = "Dummy"
	collection := db.client.Database("faryuk").Collection("group")
	filter := bson.M{"name": search}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetGroupsByNameRegex : search groups by regular expression
func (db *Handler) GetGroupsByNameRegex(search string) ([]types.Group, error) {
	var results []types.Group

	collection := db.client.Database("faryuk").Collection("group")
	filter := bson.M{"name": bson.M{"$regex": ".*" + search + ".*"}}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Group, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Group
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Group, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}
