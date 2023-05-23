package api

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"FaRyuk/internal/comment"
	"FaRyuk/internal/db"
	"FaRyuk/internal/group"
	"FaRyuk/internal/helper"
	"FaRyuk/internal/types"
	"FaRyuk/internal/user"

	"github.com/gorilla/mux"
)

const adminUsername = "admin"

func addCommentEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/get-comments/{id}", getComments).Methods("GET")
	secure.HandleFunc("/api/comments-highlights", getCommentsHighlight).Methods("GET")
	secure.HandleFunc("/api/count-highlights", getCountHighlight).Methods("GET")
	secure.HandleFunc("/api/get-tags", getTags).Methods("GET")
	secure.HandleFunc("/api/comment", insertComment).Methods("POST")
	secure.HandleFunc("/api/comment/{id}", updateComment).Methods("POST")
	secure.HandleFunc("/api/comment/{id}", deleteComment).Methods("DELETE")
}

func getComments(w http.ResponseWriter, r *http.Request) {
	var comments []types.Comment
	var result *types.Result
	var err error
	vars := mux.Vars(r)
	idResult := vars["id"]

	token, err := getCookie("jwt-token", r)
	if err != nil {
		writeForbidden(&w, "Cookie error")
		return
	}

	username, _, err := user.GetUsername(token, JWTSecret)
	if err != nil {
		writeForbidden(&w, "Token error")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	if username == adminUsername {
		comments, err = dbHandler.GetCommentsByResult(idResult)
	} else {
		result = dbHandler.GetResultByID(idResult)
		if result == nil {
			err = fmt.Errorf("cannot retreive comments for this result")
		} else {
			comments, err = dbHandler.GetCommentsByResult(idResult)
		}
	}

	if err != nil {
		writeInternalError(&w, dbError)
		return
	}
	writeObject(&w, comments)
}

func insertComment(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, "Unexpected error")
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	var content string
	err = json.Unmarshal(objmap["content"], &content)
	if err != nil {
		writeInternalError(&w, "Please provide a 'content'")
		return
	}

	var idOwner string
	err = json.Unmarshal(objmap["owner"], &idOwner)
	if err != nil {
		writeInternalError(&w, "Please provide an 'owner'")
		return
	}

	var idResult string
	err = json.Unmarshal(objmap["idResult"], &idResult)
	if err != nil {
		writeInternalError(&w, "Please provide a 'result'")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	if content[0] == '#' {
		content = strings.ToLower(content)
		err = dbHandler.AddTagsToResult(idResult, helper.GetTags(content))
	} else {
		comment := comment.NewComment(html.EscapeString(content), idOwner, idResult)
		err = dbHandler.InsertComment(comment)
	}

	if err != nil {
		writeInternalError(&w, "Database error")
		return
	}
	writeObject(&w, "Comment posted")
}

func updateComment(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, "Unexpected error")
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	var idComment string
	err = json.Unmarshal(objmap["id"], &idComment)
	if err != nil {
		writeInternalError(&w, "Please provide an 'id'")
		return
	}

	var content string
	err = json.Unmarshal(objmap["content"], &content)
	if err != nil {
		writeInternalError(&w, "Please provide a 'content'")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	comment, err := dbHandler.GetCommentByID(idComment)
	if err != nil {
		writeNotFound(&w, "Comment not found")
		return
	}

	comment.Content = html.EscapeString(content)
	comment.UpdatedDate = time.Now()
	ok := dbHandler.UpdateComment(&comment)

	if !ok {
		writeInternalError(&w, "Could not update record")
		return
	}
	writeObject(&w, "Updated successfully")
}

func deleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	res := dbHandler.RemoveCommentByID(id)
	if !res {
		writeInternalError(&w, "Could not delete record")
		return
	}
	writeObject(&w, "Deleted successfully")
}

func getCommentsHighlight(w http.ResponseWriter, r *http.Request) {
	var results []types.Result
	var err error
	pageSize := 10
	offset := 1
	search := ""

	requestedComments := make([]types.Comment, 0)
	resultsIds := make([]string, 0)
	query := r.URL.Query()

	// Check user's permission
	token, err := getCookie("jwt-token", r)
	if err != nil {
		writeForbidden(&w, "Cookie error")
		return
	}

	username, idUser, err := user.GetUsername(token, JWTSecret)
	if err != nil {
		writeForbidden(&w, "Token error")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	// Get results of current user
	if username == adminUsername {
		results, err = dbHandler.GetResultsBySearch(make(map[string]string), -1, -1)
	} else {
		user := dbHandler.GetUserByID(idUser)
		results, err = dbHandler.GetResultsBySearchAndOwner(make(map[string]string), idUser, group.ToIDsArray(user.Groups), -1, -1)
	}
	if err != nil {
		writeInternalError(&w, dbError)
		return
	}

	// Put result ids in slice
	for _, r := range results {
		resultsIds = append(resultsIds, r.ID)
	}

	searchSlice, ok := query["search"]
	if ok {
		search = searchSlice[0]
	}

	// Get all comments by chronological order
	comments, err := dbHandler.GetCommentsByText(search)
	if err != nil {
		writeInternalError(&w, dbError)
	}

	helper.Reverse(comments)

	// Check how many the user asked for
	pageSizeSlice, ok := query["size"]
	if ok {
		pageSize, _ = strconv.Atoi(pageSizeSlice[0])
	}

	offsetSlice, ok := query["offset"]
	if ok {
		offset, _ = strconv.Atoi(offsetSlice[0])
	}
	taken := 0
	for _, c := range comments {
		if len(requestedComments) >= pageSize {
			break
		}
		if !helper.ContainsStr(resultsIds, c.IDResult) {
			continue
		}
		taken++
		if taken < offset {
			continue
		}
		requestedComments = append(requestedComments, c)
	}
	writeObject(&w, requestedComments)
}

func getCountHighlight(w http.ResponseWriter, r *http.Request) {
	var results []types.Result
	var err error

	resultsIds := make([]string, 0)
	query := r.URL.Query()

	// Check user's permission
	token, err := getCookie("jwt-token", r)
	if err != nil {
		writeForbidden(&w, "Cookie error")
		return
	}

	username, idUser, err := user.GetUsername(token, JWTSecret)
	if err != nil {
		writeForbidden(&w, "Token error")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	// Get results of current user
	if username == adminUsername {
		results, err = dbHandler.GetResultsBySearch(make(map[string]string), -1, -1)
	} else {
		user := dbHandler.GetUserByID(idUser)
		results, err = dbHandler.GetResultsBySearchAndOwner(make(map[string]string), idUser, group.ToIDsArray(user.Groups), -1, -1)
	}
	if err != nil {
		writeInternalError(&w, dbError)
		return
	}

	// Put result ids in slice
	for _, r := range results {
		resultsIds = append(resultsIds, r.ID)
	}

	searchSlice, ok := query["search"]
	search := ""
	if ok {
		search = searchSlice[0]
	}

	// Get all comments by chronological order
	comments, err := dbHandler.GetCommentsByText(search)
	if err != nil {
		writeInternalError(&w, dbError)
	}

	taken := 0
	for _, c := range comments {
		if helper.ContainsStr(resultsIds, c.IDResult) {
			taken++
		}
	}
	writeObject(&w, taken)
}

func getTags(w http.ResponseWriter, r *http.Request) {
	var results []types.Result
	var err error
	tags := make(map[string]int)

	// Check user's permission
	token, err := getCookie("jwt-token", r)
	if err != nil {
		writeForbidden(&w, "Cookie error")
		return
	}

	username, idUser, err := user.GetUsername(token, JWTSecret)
	if err != nil {
		writeForbidden(&w, "Token error")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	// Get results of current user
	if username == adminUsername {
		results, err = dbHandler.GetResultsBySearch(make(map[string]string), -1, -1)
	} else {
		user := dbHandler.GetUserByID(idUser)
		results, err = dbHandler.GetResultsBySearchAndOwner(make(map[string]string), idUser, group.ToIDsArray(user.Groups), -1, -1)
	}
	if err != nil {
		writeInternalError(&w, dbError)
		return
	}

	// Put result ids in slice
	for _, r := range results {
		for _, tag := range r.Tags {
			tags[tag]++
		}
	}

	writeObject(&w, tags)
}
