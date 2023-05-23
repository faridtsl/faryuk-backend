package db

import (
	"context"
	"log"

	"FaRyuk/config"
	"FaRyuk/internal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertComment : inserts comment in the database
func (db *Handler) InsertComment(r *types.Comment) error {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	_, err := collection.InsertOne(context.TODO(), r)
	return err
}

// GetComments : gets all comments
func (db *Handler) GetComments() ([]types.Comment, error) {
	var results []types.Comment
	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.Comment, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.Comment
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Comment, 0), err
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Comment, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// RemoveCommentByID : removes comment by ID
func (db *Handler) RemoveCommentByID(id string) bool {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

// UpdateComment : updates comment
func (db *Handler) UpdateComment(r *types.Comment) bool {
	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	return err == nil
}

// GetCommentByID : retrieves comment by ID
func (db *Handler) GetCommentByID(id string) (types.Comment, error) {
	var result types.Comment
	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetCommentsByText : search comments by regular expression
func (db *Handler) GetCommentsByText(search string) ([]types.Comment, error) {
	var results []types.Comment

	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	filter := bson.M{"content": bson.M{"$regex": ".*" + search + ".*"}}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Comment, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Comment
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Comment, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// GetCommentsByTextAndOwner : searchs for comments containing a particular text and
//    that could be accessed by the current user
func (db *Handler) GetCommentsByTextAndOwner(search string, idUser string) ([]types.Comment, error) {
	var results []types.Comment

	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	filter := bson.M{"content": bson.M{"$regex": ".*" + search + ".*"}}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Comment, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Comment
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Comment, 0), err
		}
		if elem.Owner == idUser {
			results = append(results, elem)
		}
	}

	if err := cur.Err(); err != nil {
		return make([]types.Comment, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// GetCommentsByResult : get all comments for a given result
func (db *Handler) GetCommentsByResult(idResult string) ([]types.Comment, error) {
	var results []types.Comment
	collection := db.client.Database(config.Cfg.Database.Name).Collection("comment")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.Comment, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.Comment
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Comment, 0), err
		}

		if elem.IDResult == idResult {
			results = append(results, elem)
		}
	}

	if err := cur.Err(); err != nil {
		return make([]types.Comment, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}
