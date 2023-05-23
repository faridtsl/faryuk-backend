package comment_test

// TODO refactor unit tests to use mocks and add other usecases
// TODO refactor test to add setup/cleanup tests
import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"FaRyuk/api/utils"
	"FaRyuk/config"
	"FaRyuk/database/mongodb/comment"
	"FaRyuk/internal/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	db *comment.MongoRepository
	commentByIdID,
	CommentByResultID,
	commentToDeleteID,
	commentToUpdateID,
	resultID string
)

func TestMain(m *testing.M) {
	os.Setenv("CONFIGOR_ENV", "test")
	cfg, err := config.MakeConfig()
	if err != nil {
		log.Fatal("error init test config")
		os.Exit(-1)
	}
	db = comment.NewMongoRepository(cfg)
	defer db.CloseConnection()
	initIDs()
	setup()
	os.Exit(m.Run())
	// TODO add the cleanup db logic
}

func TestCreate(t *testing.T) {
	comment := &types.Comment{
		ID:          uuid.NewString(),
		Content:     "new test comment",
		Owner:       "unit test",
		CreatedDate: time.Now(),
		UpdatedDate: time.Now(),
		IDResult:    uuid.NewString(),
	}
	done := make(chan error)
	go db.Create(comment, done)
	err := <-done
	assert.NoError(t, err)
	fmt.Println(err)
}

func TestList(t *testing.T) {
	t.Run("list all comments (with no filter)", func(t *testing.T) {
		fmt.Println("\n\nlist all comments (with no filter)")
		commentsChan := make(chan *types.CommentsWithErrorType)
		go db.List(utils.ListCommentsFilter{}, commentsChan)
		result := <-commentsChan
		assert.NoError(t, result.Err)
		assert.NotEmpty(t, result.Comments)
	})

	t.Run("list comments by search text", func(t *testing.T) {
		result := make(chan *types.CommentsWithErrorType)
		search := "some-search-text"
		go db.List(utils.ListCommentsFilter{SearchText: search}, result)
		ch := <-result
		assert.NoError(t, ch.Err)
		assert.NotNil(t, ch.Comments)
	})

	t.Run("list comments by search text and owner", func(t *testing.T) {
		search := "some-search-text"
		idUser := "specefic-test-owner"
		commentsChan := make(chan *types.CommentsWithErrorType)
		go db.List(utils.ListCommentsFilter{SearchText: search, Owner: idUser}, commentsChan)
		result := <-commentsChan
		assert.NoError(t, result.Err)
		assert.NotEmpty(t, result.Comments)
	})

	t.Run("list comments by result ID", func(t *testing.T) {
		commentsChan := make(chan *types.CommentsWithErrorType)
		go db.List(utils.ListCommentsFilter{ResultID: resultID}, commentsChan)
		result := <-commentsChan
		assert.NoError(t, result.Err)
		assert.NotEmpty(t, result.Comments)
	})
}

func TestDelete(t *testing.T) {
	done := make(chan error)
	go db.Delete(commentToDeleteID, done)
	err := <-done
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	done := make(chan error)
	comment := &types.Comment{
		ID:          commentToUpdateID,
		Content:     "updated comment",
		Owner:       "unit test",
		UpdatedDate: time.Now(),
		IDResult:    resultID,
	}

	go db.Update(comment, done)
	err := <-done
	assert.NoError(t, err)
	// TODO get commlent by id and check if it is correctly updated
}

func TestGetByID(t *testing.T) {
	result := make(chan *types.CommentWithErrorType)
	go db.GetByID(commentByIdID, result)
	ch := <-result
	assert.NotNil(t, ch.Comment)
	assert.NoError(t, ch.Err)
}

func setup() error {
	// Set up test configuration
	os.Setenv("CONFIGOR_ENV", "test")
	for _, testData := range getCommentsTestData() {
		done := make(chan error)

		go db.Create(&testData, done)
		err := <-done
		if err != nil {
			return errors.New("element not inserted")
		}
	}
	return nil
}

func getCommentsTestData() []types.Comment {
	return []types.Comment{
		{
			ID:          commentByIdID,
			Content:     "new test mockComment",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    uuid.NewString(),
		},
		{
			ID:          commentToDeleteID,
			Content:     "some comment to be deleted",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    uuid.NewString(),
		},
		{
			ID:          commentToUpdateID,
			Content:     "test comment to update",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    uuid.NewString(),
		},
		{
			ID:          CommentByResultID,
			Content:     "new test comment to be filtered by result ID",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    resultID,
		},
		{
			ID:          commentToUpdateID,
			Content:     "new test comment to be updated",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    resultID,
		},
		{
			ID:          uuid.NewString(),
			Content:     "some-search-text",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    uuid.NewString(),
		},
		{
			ID:          uuid.NewString(),
			Content:     "some-search-text",
			Owner:       "specefic-test-owner",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    uuid.NewString(),
		},
	}
}

func initIDs() {
	commentByIdID = uuid.NewString()
	CommentByResultID = uuid.NewString()
	commentToDeleteID = uuid.NewString()
	commentToUpdateID = uuid.NewString()
	resultID = uuid.NewString()
}
