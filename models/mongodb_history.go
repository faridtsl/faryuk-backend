package models

import (
	"context"
	"log"
	"strconv"

	"FaRyuk/internal/helper"
	"FaRyuk/internal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertHistoryRecord : inserts history record in the database
func (db *Handler) InsertHistoryRecord(r types.HistoryRecord) error {
	collection := db.client.Database("faryuk").Collection("history")

	_, err := collection.InsertOne(context.TODO(), r)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// GetHistoryRecords : returns all history records
func (db *Handler) GetHistoryRecords() []types.HistoryRecord {
	var results []types.HistoryRecord
	collection := db.client.Database("faryuk").Collection("history")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
		return make([]types.HistoryRecord, 0)
	}

	for cur.Next(context.TODO()) {
		var elem types.HistoryRecord
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
			return make([]types.HistoryRecord, 0)
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
		return make([]types.HistoryRecord, 0)
	}

	cur.Close(context.TODO())
	helper.Reverse(results)
	return results
}

// RemoveHistoryRecordByID : removes a history record by ID
func (db *Handler) RemoveHistoryRecordByID(id string) bool {
	collection := db.client.Database("faryuk").Collection("history")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

// UpdateHistoryRecord : updates a history record
func (db *Handler) UpdateHistoryRecord(r types.HistoryRecord) bool {
	collection := db.client.Database("faryuk").Collection("history")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

// GetHistoryRecordByID : returns one history record by ID
func (db *Handler) GetHistoryRecordByID(id string) (types.HistoryRecord, error) {
	var result types.HistoryRecord
	collection := db.client.Database("faryuk").Collection("history")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetHistoryRecordsBySearch : returns history records by search criteria
func (db *Handler) GetHistoryRecordsBySearch(search map[string]string,
	offset int,
	pageSize int) ([]types.HistoryRecord, error) {
	var results []types.HistoryRecord
	var opts options.FindOptions

	collection := db.client.Database("faryuk").Collection("history")
	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"state":      bson.M{"$regex": ".*" + search["state"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
	}

	if search["isFinished"] != "" {
		isFinished, err := strconv.ParseBool(search["isFinished"])
		if err == nil {
			filter["isFinished"] = isFinished
		}
	}

	if search["isSuccess"] != "" {
		isSuccess, err := strconv.ParseBool(search["isSuccess"])
		if err == nil {
			filter["isSuccess"] = isSuccess
		}
	}

	if search["isWeb"] != "" {
		isWeb, err := strconv.ParseBool(search["isWeb"])
		if err == nil {
			filter["isWeb"] = isWeb
		}
	}

	skip := int64(offset)
	limit := int64(pageSize)

	if limit == -1 {
		opts = options.FindOptions{}
	} else {
		opts = options.FindOptions{
			Skip:  &skip,
			Limit: &limit,
		}
	}

	opts.SetSort(bson.M{"$natural": -1})

	cur, err := collection.Find(context.TODO(), filter, &opts)
	if err != nil {
		return make([]types.HistoryRecord, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.HistoryRecord
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.HistoryRecord, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// GetHistoryRecordsBySearchAndOwner : returns history records by search criteria for a given owner
func (db *Handler) GetHistoryRecordsBySearchAndOwner(search map[string]string,
	idUser string,
	groups []string,
	offset int,
	pageSize int) ([]types.HistoryRecord, error) {
	var results []types.HistoryRecord
	var opts options.FindOptions

	collection := db.client.Database("faryuk").Collection("history")
	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"state":      bson.M{"$regex": ".*" + search["state"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
		"$or": []interface{}{
			bson.M{"owner": idUser},
			bson.M{"sharedWith": idUser},
			bson.M{"ownerGroup": bson.M{"$in": groups}},
		},
	}

	if search["isFinished"] != "" {
		isFinished, err := strconv.ParseBool(search["isFinished"])
		if err == nil {
			filter["isFinished"] = isFinished
		}
	}

	if search["isSuccess"] != "" {
		isSuccess, err := strconv.ParseBool(search["isSuccess"])
		if err == nil {
			filter["isSuccess"] = isSuccess
		}
	}

	if search["isWeb"] != "" {
		isWeb, err := strconv.ParseBool(search["isWeb"])
		if err == nil {
			filter["isWeb"] = isWeb
		}
	}

	skip := int64(offset)
	limit := int64(pageSize)

	if limit == -1 {
		opts = options.FindOptions{}
	} else {
		opts = options.FindOptions{
			Skip:  &skip,
			Limit: &limit,
		}
	}

	opts.SetSort(bson.M{"$natural": -1})

	cur, err := collection.Find(context.TODO(), filter, &opts)
	if err != nil {
		return make([]types.HistoryRecord, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.HistoryRecord
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.HistoryRecord, 0), err
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.HistoryRecord, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// CountHistoryRecordsBySearch : returns history records by search criteria for a given owner
func (db *Handler) CountHistoryRecordsBySearch(search map[string]string) (int, error) {
	collection := db.client.Database("faryuk").Collection("history")
	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"state":      bson.M{"$regex": ".*" + search["state"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
	}

	if search["isFinished"] != "" {
		isFinished, err := strconv.ParseBool(search["isFinished"])
		if err == nil {
			filter["isFinished"] = isFinished
		}
	}

	if search["isSuccess"] != "" {
		isSuccess, err := strconv.ParseBool(search["isSuccess"])
		if err == nil {
			filter["isSuccess"] = isSuccess
		}
	}

	if search["isWeb"] != "" {
		isWeb, err := strconv.ParseBool(search["isWeb"])
		if err == nil {
			filter["isWeb"] = isWeb
		}
	}

	cnt, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return -1, err
	}
	return int(cnt), nil
}

// CountHistoryRecordsBySearchAndOwner : returns history records by search criteria for a given owner
func (db *Handler) CountHistoryRecordsBySearchAndOwner(search map[string]string, groups []string,
	idUser string) (int, error) {
	collection := db.client.Database("faryuk").Collection("history")
	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"state":      bson.M{"$regex": ".*" + search["state"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
		"$or": []interface{}{
			bson.M{"owner": idUser},
			bson.M{"sharedWith": idUser},
			bson.M{"ownerGroup": bson.M{"$in": groups}},
		},
	}

	if search["isFinished"] != "" {
		isFinished, err := strconv.ParseBool(search["isFinished"])
		if err == nil {
			filter["isFinished"] = isFinished
		}
	}

	if search["isSuccess"] != "" {
		isSuccess, err := strconv.ParseBool(search["isSuccess"])
		if err == nil {
			filter["isSuccess"] = isSuccess
		}
	}

	if search["isWeb"] != "" {
		isWeb, err := strconv.ParseBool(search["isWeb"])
		if err == nil {
			filter["isWeb"] = isWeb
		}
	}

	cnt, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return -1, err
	}
	return int(cnt), nil
}

// GetHistoryRecordsByOwner : returns all history records that a given owner can access
func (db *Handler) GetHistoryRecordsByOwner(idUser string) ([]types.HistoryRecord, error) {
	var results []types.HistoryRecord
	collection := db.client.Database("faryuk").Collection("history")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.HistoryRecord, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.HistoryRecord
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.HistoryRecord, 0), err
		}

		if elem.Owner == idUser {
			results = append(results, elem)
		}
	}

	if err := cur.Err(); err != nil {
		return make([]types.HistoryRecord, 0), err
	}

	cur.Close(context.TODO())
	helper.Reverse(results)
	return results, nil
}
