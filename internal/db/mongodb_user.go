package db

import (
	"FaRyuk/config"
	"FaRyuk/internal/types"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertUser : inserts user in database
func (db *Handler) InsertUser(r *types.User) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	_, err := collection.InsertOne(context.TODO(), r)
	return err
}

// GetUsers : returns all users
func (db *Handler) GetUsers() []types.User {
	var users []types.User
	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.User, 0)
	}

	for cur.Next(context.TODO()) {
		var elem types.User
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		users = append(users, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
		return make([]types.User, 0)
	}

	cur.Close(context.TODO())
	return users
}

// RemoveUserByID : removes a user by its ID
func (db *Handler) RemoveUserByID(id string) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	return err
}

// UpdateUser : updates a user
func (db *Handler) UpdateUser(r *types.User) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	return err
}

// GetUserByID : gets a user by its ID
func (db *Handler) GetUserByID(id string) *types.User {
	var user types.User
	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&user)
	if err != nil {
		return nil
	}
	return &user
}

// GetUserByUsername : returns a user by its username
func (db *Handler) GetUserByUsername(username string) *types.User {
	var user types.User
	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil
	}
	return &user
}

// GetUsersByGroup : returns all users in given group
func (db *Handler) GetUsersByGroup(group types.Group) ([]types.User, error) {
	var opts options.FindOptions
	var users []types.User
	groups := make([]types.Group, 0)
	groups = append(groups, group)

	collection := db.client.Database(config.Cfg.Database.Name).Collection("users")
	filter := bson.M{"groups": bson.M{"$all": groups}}

	cur, err := collection.Find(context.TODO(), filter, &opts)
	if err != nil {
		return make([]types.User, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.User
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		elem.Groups = groups
		users = append(users, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.User, 0), err
	}

	cur.Close(context.TODO())
	return users, nil

}
