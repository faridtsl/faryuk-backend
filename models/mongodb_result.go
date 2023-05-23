package models

import (
	"context"
	"fmt"
	"log"
	"strings"

	"FaRyuk/internal/helper"
	"FaRyuk/internal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const emptyResult = "empty"

// InsertResult : inserts result in the database
func (db *Handler) InsertResult(r *types.Result) error {
	collection := db.client.Database("faryuk").Collection("results")

	_, err := collection.InsertOne(context.TODO(), r)
	return err
}

// GetResults : returns all results from database
func (db *Handler) GetResults() []types.Result {
	var results []types.Result
	collection := db.client.Database("faryuk").Collection("results")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
		return make([]types.Result, 0)
	}

	for cur.Next(context.TODO()) {
		var elem types.Result
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
		return make([]types.Result, 0)
	}

	cur.Close(context.TODO())
	helper.Reverse(results)
	return results
}

// RemoveByID : removes a result by ID
func (db *Handler) RemoveByID(id string) bool {
	collection := db.client.Database("faryuk").Collection("results")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"id": id})
	return err == nil
}

// UpdateResult : updates result
func (db *Handler) UpdateResult(r *types.Result) bool {
	collection := db.client.Database("faryuk").Collection("results")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"id": r.ID}, bson.M{"$set": r})
	return err == nil
}

// GetResultByID : returns a result by ID
func (db *Handler) GetResultByID(id string) *types.Result {
	var result types.Result
	collection := db.client.Database("faryuk").Collection("results")
	err := collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if err != nil {
		return nil
	}
	return &result
}

// GetResultsBySearch : returns all results matching search criteria
func (db *Handler) GetResultsBySearch(search map[string]string, offset, pageSize int) ([]types.Result, error) {
	var results []types.Result
	var opts options.FindOptions

	collection := db.client.Database("faryuk").Collection("results")
	filter := bson.M{
		"host":       bson.M{"$regex": ".*" + search["default"] + ".*"},
		"ips":        bson.M{"$regex": ".*" + search["ip"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
	}
	if len(helper.ParseInts(search["ports"])) != 0 {
		if search["ports"] != emptyResult {
			filter["openPorts"] = bson.M{"$all": helper.ParseInts(search["ports"])}
		}
	} else {
		filter["openPorts"] = bson.M{"$ne": make([]int, 0)}
	}

	if search["buster"] != "" {
		filter["webResults.busterres.path"] = bson.M{"$regex": ".*" + search["buster"] + ".*"}
	}

	if search["tags"] != "" {
		searchTags := make([]string, 0)
		for _, tag := range strings.Split(search["tags"], ",") {
			searchTags = append(searchTags, "#"+tag)
		}
		filter["tags"] = bson.M{"$all": searchTags}
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
		return make([]types.Result, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.Result
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Result, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// GetResultsBySearchAndOwner : returns all results matching search criteria and that a user can access
func (db *Handler) GetResultsBySearchAndOwner(search map[string]string,
	idUser string,
	groups []string,
	offset int,
	pageSize int) ([]types.Result, error) {
	var results []types.Result
	var opts options.FindOptions

	collection := db.client.Database("faryuk").Collection("results")

	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"ips":        bson.M{"$regex": ".*" + search["ip"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
		"$or": []interface{}{
			bson.M{"owner": idUser},
			bson.M{"sharedWith": idUser},
			bson.M{"ownerGroup": bson.M{"$in": groups}},
		},
	}
	if len(helper.ParseInts(search["ports"])) != 0 {
		if search["ports"] != emptyResult {
			filter["openPorts"] = bson.M{"$all": helper.ParseInts(search["ports"])}
		}
	} else {
		filter["openPorts"] = bson.M{"$ne": make([]int, 0)}
	}

	if search["buster"] != "" {
		filter["webResults.busterres.path"] = bson.M{"$regex": ".*" + search["buster"] + ".*"}
	}

	if search["tags"] != "" {
		searchTags := make([]string, 0)
		for _, tag := range strings.Split(search["tags"], ",") {
			searchTags = append(searchTags, "#"+tag)
		}
		filter["tags"] = bson.M{"$all": searchTags}
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
		return make([]types.Result, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Result
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Result, 0), err
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Result, 0), err
	}

	cur.Close(context.TODO())
	return results, nil
}

// CountResultsBySearch : returns all results matching search criteria and that a user can access
func (db *Handler) CountResultsBySearch(search map[string]string) (int, error) {
	collection := db.client.Database("faryuk").Collection("results")
	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"ips":        bson.M{"$regex": ".*" + search["ip"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
	}
	if len(helper.ParseInts(search["ports"])) != 0 {
		if search["ports"] != emptyResult {
			filter["openPorts"] = bson.M{"$all": helper.ParseInts(search["ports"])}
		}
	} else {
		filter["openPorts"] = bson.M{"$ne": make([]int, 0)}
	}

	if search["buster"] != "" {
		filter["webResults.busterres.path"] = bson.M{"$regex": ".*" + search["buster"] + ".*"}
	}

	if search["tags"] != "" {
		searchTags := make([]string, 0)
		for _, tag := range strings.Split(search["tags"], ",") {
			searchTags = append(searchTags, "#"+tag)
		}
		filter["tags"] = bson.M{"$all": searchTags}
	}

	cnt, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return -1, err
	}
	return int(cnt), nil
}

// CountResultsBySearchAndOwner : returns all results matching search criteria and that a user can access
func (db *Handler) CountResultsBySearchAndOwner(search map[string]string, groups []string,
	idUser string) (int, error) {
	collection := db.client.Database("faryuk").Collection("results")
	filter := bson.M{"host": bson.M{"$regex": ".*" + search["default"] + ".*"},
		"ips":        bson.M{"$regex": ".*" + search["ip"] + ".*"},
		"ownerGroup": bson.M{"$regex": ".*" + search["group"] + ".*"},
		"$or": []interface{}{
			bson.M{"owner": idUser},
			bson.M{"sharedWith": idUser},
			bson.M{"ownerGroup": bson.M{"$in": groups}},
		},
	}
	if len(helper.ParseInts(search["ports"])) != 0 {
		if search["ports"] != emptyResult {
			filter["openPorts"] = bson.M{"$all": helper.ParseInts(search["ports"])}
		}
	} else {
		filter["openPorts"] = bson.M{"$ne": make([]int, 0)}
	}

	if search["buster"] != "" {
		filter["webResults.busterres.path"] = bson.M{"$regex": ".*" + search["buster"] + ".*"}
	}

	if search["tags"] != "" {
		searchTags := make([]string, 0)
		for _, tag := range strings.Split(search["tags"], ",") {
			searchTags = append(searchTags, "#"+tag)
		}
		filter["tags"] = bson.M{"$all": searchTags}
	}

	cnt, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return -1, err
	}
	return int(cnt), nil
}

// GetResultsByHostAndOwner : returns all results matching search host and that a user can access
func (db *Handler) GetResultsByHostAndOwner(search, idUser string) ([]types.Result, error) {
	var results []types.Result

	collection := db.client.Database("faryuk").Collection("results")
	filter := bson.M{"host": search}

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return make([]types.Result, 0), err
	}
	for cur.Next(context.TODO()) {
		var elem types.Result
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Result, 0), err
		}
		if elem.Owner != idUser && !helper.ContainsStr(elem.SharedWith, idUser) {
			continue
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return make([]types.Result, 0), err
	}

	cur.Close(context.TODO())
	helper.Reverse(results)
	return results, nil
}

// GetResultsByOwner : returns all results that a user can access
func (db *Handler) GetResultsByOwner(idUser string) ([]types.Result, error) {
	var results []types.Result
	collection := db.client.Database("faryuk").Collection("results")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return make([]types.Result, 0), err
	}

	for cur.Next(context.TODO()) {
		var elem types.Result
		err := cur.Decode(&elem)
		if err != nil {
			return make([]types.Result, 0), err
		}

		if elem.Owner == idUser || helper.ContainsStr(elem.SharedWith, idUser) {
			results = append(results, elem)
		}
	}

	if err := cur.Err(); err != nil {
		return make([]types.Result, 0), err
	}

	cur.Close(context.TODO())
	helper.Reverse(results)
	return results, nil
}

// AddTagsToResult : Add a tag to result if it does not exist
func (db *Handler) AddTagsToResult(idResult string, tags []string) error {
	result := db.GetResultByID(idResult)
	if result == nil {
		return fmt.Errorf("no result with such id")
	}
	for _, tag := range tags {
		if !helper.ContainsStr(result.Tags, tag) {
			result.Tags = append(result.Tags, tag)
		}
	}
	ok := db.UpdateResult(result)
	if !ok {
		return fmt.Errorf("could not update result")
	}
	return nil
}
