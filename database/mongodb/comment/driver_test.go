package comment_test

// TODO refactor unit tests to use mocks and add other usecases
// TODO refactor test to add setup/cleanup tests
import (
	"fmt"
	"os"
	"testing"
	"time"

	"FaRyuk/config"
	"FaRyuk/database/mongodb/comment"
	"FaRyuk/internal/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	db *comment.MongoCommentRepository
)

func TestMain(m *testing.M) {
	os.Setenv("CONFIGOR_ENV", "test")
	cfg, _ := config.MakeConfig()
	db = comment.NewMongoCommentRepository(cfg)
	
	// if !err {
	// 	fmt.Println(err)
	// 	os.Exit(-1)
	// }
	os.Exit(m.Run())
	// TODO add the cleanup db logic on a defer
}

func TestInsertComment(t *testing.T) {
	comment := &types.Comment{
		ID:          uuid.NewString(),
		Content:     "new test comment",
		Owner:       "unit test",
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
		IDResult:    uuid.NewString(),
	}
	done := make(chan bool)
	go db.InsertComment(comment, done)
	err := <-done
	assert.True(t, err)
}

func TestGetComments(t *testing.T) {
	commentsChan := make(chan types.CommentsWithErrorType)
	go db.GetComments(commentsChan)
	result := <-commentsChan
	assert.NoError(t, result.Err)
	assert.NotEmpty(t, result.Comments)
	fmt.Println(result.Comments)
}

func TestRemoveCommentByID(t *testing.T) {
	done := make(chan error)
	id := "some-id"
	go db.RemoveCommentByID(id, done)
	err := <-done
	assert.NoError(t, err)
}

func TestUpdateComment(t *testing.T) {
	commentToUpdate := &types.Comment{
		ID:          uuid.NewString(),
		Content:     "test comment to update",
		Owner:       "unit test",
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
		IDResult:    uuid.NewString(),
	}
	insertDone := make(chan bool)
	go db.InsertComment(commentToUpdate, insertDone)
	err := <-insertDone
	assert.True(t, err)
	done := make(chan bool)
	comment := &types.Comment{
		ID:          uuid.NewString(),
		Content:     "updated comment",
		Owner:       "unit test",
		UpdatedDate: time.Now(),
		IDResult:    commentToUpdate.IDResult,
	}

	go db.UpdateComment(comment, done)
	result := <-done
	assert.True(t, result)
	// TODO get commlent by id and check if it is correctly updated
}

func TestGetCommentByID(t *testing.T) {
	comment := &types.Comment{
		ID:       "some-id",
		Content:  "some-search-text",
		IDResult: "some-result-id",
		Owner:    "owner-1",
	}
	done := make(chan bool)
	go db.InsertComment(comment, done)
	err := <-done
	assert.True(t, err)

	result := make(chan types.CommentWithErrorType)
	id := "some-id"
	go db.GetCommentByID(id, result)
	ch := <-result
	assert.NotNil(t, ch.Comment)
	assert.NoError(t, ch.Err)
}

func TestGetCommentsByText(t *testing.T) {
	result := make(chan types.CommentsWithErrorType)
	search := "some-search-text"
	go db.GetCommentsByText(search, result)
	ch := <-result
	assert.NotNil(t, ch.Comments)
	assert.NoError(t, ch.Err)
}

func TestGetCommentsByTextAndOwner(t *testing.T) {
	search := "some-search-text"
	idUser := "owner-1"
	commentsChan := make(chan types.CommentsWithErrorType)
	go db.GetCommentsByTextAndOwner(search, idUser, commentsChan)
	result := <-commentsChan
	assert.NoError(t, result.Err)
	assert.NotEmpty(t, result.Comments)
}

func TestGetCommentsByResult(t *testing.T) {
	commentsChan := make(chan types.CommentsWithErrorType)
	idResult := "some-result-id"
	go db.GetCommentsByResultID(idResult, commentsChan)
	result := <-commentsChan
	assert.NoError(t, result.Err)
	assert.NotEmpty(t, result.Comments)
}