package comment

import (
	"FaRyuk/api/utils"
	"FaRyuk/config"
	"FaRyuk/database"
	"FaRyuk/internal/types"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TODO move to new file: API mapper
func AddCommentEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/v1/comments", ListComments).Methods(http.MethodGet)
	secure.HandleFunc("/api/v1/comments/{id}", DeleteComment).Methods(http.MethodDelete)
	secure.HandleFunc("/api/v1/comments/{id}", GetCommentByID).Methods(http.MethodGet)
	secure.HandleFunc("/api/v1/comments", CreateComment).Methods(http.MethodPost)
	secure.HandleFunc("/api/v1/comments/{id}", UpdateComment).Methods(http.MethodPut)
}

func ListComments(w http.ResponseWriter, rr *http.Request) {
	cfg, confErr := config.MakeConfig()
	if confErr != nil {
		log.Fatal(confErr)
	}

	dbHandler, err := database.CreateDbHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer dbHandler.CloseConnection()

	listCommentsFilter := extractFilter(rr)
	commentsChan := make(chan *types.CommentsWithErrorType)

	go dbHandler.List(listCommentsFilter, commentsChan)

	result := <-commentsChan
	if result.Err != nil {
		utils.WriteInternalError(&w, result.Err.Error())
		return
	}

	utils.ReturnSuccess(&w, result.Comments)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]

	dbHandler := createDbHandler()
	defer dbHandler.CloseConnection()

	done := make(chan error)

	go dbHandler.Delete(commentID, done)
	err := <-done
	if err != nil {
		utils.WriteInternalError(&w, err.Error())
		return
	}

	utils.ReturnSuccessNoContent(&w)
}

func GetCommentByID(w http.ResponseWriter, r *http.Request) {
	// Get the comment ID from the request URL parameters
	vars := mux.Vars(r)
	commentID := vars["id"]

	dbHandler := createDbHandler()
	defer dbHandler.CloseConnection()

	// Create a channel to receive the result of the database operation
	commentChan := make(chan *types.CommentWithErrorType)

	// Call the Delete method asynchronously
	go dbHandler.GetByID(commentID, commentChan)

	// Wait for the database operation to complete and check for errors
	result := <-commentChan

	if result.Err != nil {
		// If an error occurred, return an appropriate HTTP response
		utils.WriteInternalError(&w, result.Err.Error())
		return
	}

	if result.Comment == nil {
		// If an error occurred, return an appropriate HTTP response
		utils.WriteNotFound(&w)
		return
	}

	// If the operation was successful, return a success response
	utils.ReturnSuccess(&w, nil)
}

func CreateComment(w http.ResponseWriter, r *http.Request) {
	var newComment types.Comment

	err := json.NewDecoder(r.Body).Decode(&newComment)
	if err != nil {
		utils.WriteBadRequest(&w)
	}

	dbHandler := createDbHandler()
	defer dbHandler.CloseConnection()

	done := make(chan error)
	go dbHandler.Create(&newComment, done)
	err = <-done

	if err != nil {
		utils.WriteInternalError(&w, err.Error())
		return
	}
	utils.ReturnSuccessCreated(&w)
}

func UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	var newComment types.Comment

	err := json.NewDecoder(r.Body).Decode(&newComment)
	if err != nil {
		utils.WriteBadRequest(&w)
	}
	newComment.ID = commentID

	dbHandler := createDbHandler()
	defer dbHandler.CloseConnection()

	done := make(chan error)
	go dbHandler.Update(&newComment, done)
	err = <-done

	if err != nil {
		utils.WriteInternalError(&w, err.Error())
		return
	}

	utils.ReturnSuccessCreated(&w)
}

func createDbHandler() database.DbHandler {
	cfg, confErr := config.MakeConfig()
	if confErr != nil {
		log.Fatal(confErr)
	}
	dbHandler, err := database.CreateDbHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return dbHandler
}

func extractFilter(rr *http.Request) utils.ListCommentsFilter {
	resultID := rr.URL.Query().Get("resultID")
	searchText := rr.URL.Query().Get("searchText")
	owner := rr.URL.Query().Get("owner")
	listCommentsFilter := utils.ListCommentsFilter{
		ResultID:   resultID,
		SearchText: searchText,
		Owner:      owner,
	}
	return listCommentsFilter
}
