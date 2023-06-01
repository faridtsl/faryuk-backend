package comment_test

import (
	comment_api "FaRyuk/api/comment"
	"FaRyuk/config"
	"FaRyuk/internal/types"
	"errors"
	"fmt"
	"log"
	"time"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"FaRyuk/database/mongodb/comment"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	db                                                            *comment.MongoRepository
	commentByIdID, CommentByResultID, commentToDeleteID, resultID string
)

func TestMain(m *testing.M) {

	os.Setenv("CONFIGOR_ENV", "test")
	cfg, err := config.MakeConfig()
	if err != nil {
		log.Fatal("error init test config")
		os.Exit(-1)
	}
	db = comment.NewMongoRepository(cfg)
	defer db.CloseCommentDBConnection()

	initIDs()
	err = setup()
	if err != nil {
		log.Fatal("error inserting test data")
		os.Exit(-1)
	}

	// Run the tests
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestDelete(t *testing.T) {
	commentID := "some-id"

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/comments:%s", commentID), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	comment_api.DeleteComment(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestListComments(t *testing.T) {
	t.Run("list all comments (with no filter)", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/comments", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		comment_api.ListComments(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		var response struct {
			Body []types.Comment `json:"body"`
		}

		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		comments := response.Body
		assert.NotEmpty(t, comments)

		for _, mockComment := range comments {
			if !(mockComment.Owner == "unit test" || mockComment.Owner == "specefic-test-owner") {
				t.Fatal("wrong comment received", mockComment.Owner)
			}
			assert.True(t, mockComment.Content != "")
			assert.True(t, mockComment.IDResult != "")
		}
	})

	t.Run("list comments by search text", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("/comments?searchText=%s", "some-search-text"), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		comment_api.ListComments(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotNil(t, rr.Body)
		var comments []types.Comment
		var response struct {
			Body []types.Comment `json:"body"`
		}

		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}
		if len(response.Body) == 0 {
			t.Fail()
		}
		for _, com := range comments {
			if com.Content != "some-search-text" {
				t.Fail()
			}
		}
	})

	t.Run("list comments by search text and owner", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("/comments?searchText=%s&owner=%s", "some-search-text", "specefic-test-owner"), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		comment_api.ListComments(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotNil(t, rr.Body)
		var comments []types.Comment
		var response struct {
			Body []types.Comment `json:"body"`
		}

		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}
		if len(response.Body) == 0 {
			t.Fail()
		}
		for _, com := range comments {
			if com.Content != "some-search-text" {
				t.Fail()
			}
			if com.Content != "specefic-owner" {
				t.Fail()
			}
		}
	})

	t.Run("list comments by resultID", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("/comments?resultID=%s", resultID), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		comment_api.ListComments(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotNil(t, rr.Body)
		var comments []types.Comment
		var response struct {
			Body []types.Comment `json:"body"`
		}

		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}
		if len(response.Body) == 0 {
			t.Fail()
		}
		for _, com := range comments {
			if com.Content != "new test comment to be filtered by result ID" {
				t.Fail()
			}
		}
	})
}

func setup() error {
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
			ID:          CommentByResultID,
			Content:     "new test comment to be filtered by result ID",
			Owner:       "unit test",
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
			IDResult:    resultID,
		},
	}
}

func initIDs() {
	commentByIdID = uuid.NewString()
	CommentByResultID = uuid.NewString()
	commentToDeleteID = uuid.NewString()
	resultID = uuid.NewString()
}
