package comment

import (
	"FaRyuk/api/utils"
	"FaRyuk/config"
	"FaRyuk/internal/types"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
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
func (repo *MongoRepository) CloseConnection() {
	err := repo.Database().Client().Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
		return
	}
}

// Create : creates a new comment in the database
func (repo *MongoRepository) Create(comment *types.Comment, done chan<- error) {
	defer close(done)
	if comment.ID == "" {
		comment.ID = uuid.NewString() // Generate a new UUID for the comment
	}
	_, err := repo.InsertOne(context.TODO(), comment)
	done <- err
}

// Delete : removes comment by ID
func (repo *MongoRepository) Delete(id string, done chan<- error) {
	_, err := repo.DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		done <- err
		return
	}
	done <- nil
}

// Update : updates comment
func (repo *MongoRepository) Update(r *types.Comment, done chan<- error) {
	_, err := repo.UpdateOne(context.Background(), bson.M{"id": r.ID}, bson.M{"$set": r})
	done <- err
}

// GetByID : retrieves comment by ID
func (repo *MongoRepository) GetByID(id string, result chan<- *types.CommentWithErrorType) {
	var comment *types.Comment
	err := repo.FindOne(context.Background(), bson.M{"id": id}).Decode(&comment)
	if err != nil {
		result <- &types.CommentWithErrorType{Comment: &types.Comment{}, Err: err}
	} else {
		result <- &types.CommentWithErrorType{Comment: comment, Err: nil}
	}
}

// List : gets all comments (A filter is supprted)
func (repo *MongoRepository) List(listCommentsFilter utils.ListCommentsFilter, commentsChan chan<- *types.CommentsWithErrorType) {
	defer close(commentsChan)
	var results []*types.Comment

	cur, err := repo.Find(context.TODO(), buildFilter(listCommentsFilter), options.Find())
	if err != nil {
		commentsChan <- &types.CommentsWithErrorType{Comments: nil, Err: err}
		return
	}
	if err := cur.Err(); err != nil {
		commentsChan <- &types.CommentsWithErrorType{Comments: nil, Err: err}
		return
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var elem *types.Comment
		err := cur.Decode(&elem)
		if err != nil {
			commentsChan <- &types.CommentsWithErrorType{Comments: nil, Err: err}
			return
		}
		results = append(results, elem)
	}

	commentsChan <- &types.CommentsWithErrorType{Comments: results, Err: nil}
}

func buildFilter(listCommentsFilter utils.ListCommentsFilter) interface{} {
	if listCommentsFilter.Owner == "" &&
		listCommentsFilter.SearchText == "" &&
		listCommentsFilter.ResultID == "" {
		return bson.D{}
	}

	filters := []bson.M{}

	if listCommentsFilter.Owner != "" {
		filters = append(filters, bson.M{"owner": listCommentsFilter.Owner})
	}
	if listCommentsFilter.SearchText != "" {
		filters = append(filters, bson.M{"content": listCommentsFilter.SearchText})
	}
	if listCommentsFilter.ResultID != "" {
		filters = append(filters, bson.M{"idResult": listCommentsFilter.ResultID})
	}

	return bson.M{"$and": filters}
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
