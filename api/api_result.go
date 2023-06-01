package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"FaRyuk/api/utils"
	"FaRyuk/internal/group"
	"FaRyuk/internal/helper"
	"FaRyuk/internal/types"
	"FaRyuk/models"

	"github.com/gorilla/mux"
)

func addResultEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/get-results", getResults).Methods("GET")
	secure.HandleFunc("/api/count-results", countResults).Methods("GET")
	secure.HandleFunc("/api/result/{id}", getResultByID).Methods("GET")
	secure.HandleFunc("/api/result/{id}", deleteResultByID).Methods("DELETE")
	secure.HandleFunc("/api/delete-tag", deleteTag).Methods("POST")
}

func getResults(w http.ResponseWriter, r *http.Request) {
	var results []types.Result
	var err error
	var pageSize int
	var offset int

	query := r.URL.Query()
	searchSlice, ok := query["search"]
	search := ""
	if ok {
		search = searchSlice[0] + " "
	}

	searchMap := helper.Tokenize(search)

	username, idUser, err := getIdentity(&w, r)
	if err != nil {
		return
	}

	pageSizeSlice, ok := query["size"]
	pageSize = 10
	if ok {
		pageSize, _ = strconv.Atoi(pageSizeSlice[0])
	}

	offsetSlice, ok := query["offset"]
	offset = 1
	if ok {
		offset, _ = strconv.Atoi(offsetSlice[0])
	}

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseCommentDBConnection()

	if searchMap["group"] != "" {
		group, err := dbHandler.GetGroupsByName(searchMap["group"])
		if err != nil {
			utils.WriteInternalError(&w, dbError)
			return
		}
		searchMap["group"] = group.ID
	}
	if username == "admin" {
		results, err = dbHandler.GetResultsBySearch(searchMap, offset, pageSize)
	} else {
		user := dbHandler.GetUserByID(idUser)
		results, err = dbHandler.GetResultsBySearchAndOwner(searchMap, idUser, group.ToIDsArray(user.Groups), offset, pageSize)
	}

	if err != nil {
		utils.WriteInternalError(&w, dbError)
		return
	}
	if len(results) == 0 {
		utils.ReturnSuccess(&w, results)
		return
	}

	utils.ReturnSuccess(&w, results)
}

func countResults(w http.ResponseWriter, r *http.Request) {
	var cntResults int
	var err error
	query := r.URL.Query()
	searchSlice, ok := query["search"]
	search := ""
	if ok {
		search = searchSlice[0] + " "
	}

	searchMap := helper.Tokenize(search)

	username, idUser, err := getIdentity(&w, r)
	if err != nil {
		return
	}

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseCommentDBConnection()

	if username == "admin" {
		cntResults, err = dbHandler.CountResultsBySearch(searchMap)
	} else {
		user := dbHandler.GetUserByID(idUser)
		cntResults, err = dbHandler.CountResultsBySearchAndOwner(searchMap, group.ToIDsArray(user.Groups), idUser)
	}
	if err != nil {
		utils.WriteInternalError(&w, dbError)
		return
	}
	utils.ReturnSuccess(&w, cntResults)
}

func getResultByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseCommentDBConnection()

	result := dbHandler.GetResultByID(id)
	if result == nil {
		utils.WriteInternalError(&w, dbError)
		return
	}

	utils.ReturnSuccess(&w, *result)
}

func deleteResultByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseCommentDBConnection()

	result := dbHandler.GetResultByID(id)
	if result == nil {
		utils.WriteInternalError(&w, dbError)
		return
	}

	username, idUser, err := getIdentity(&w, r)
	if err != nil || (username != "admin" && idUser != result.Owner) {
		return
	}

	res := dbHandler.RemoveByID(id)
	if !res {
		utils.WriteInternalError(&w, dbError)
		return
	}

	utils.ReturnSuccess(&w, "Deleted successfully")
}

func deleteTag(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage
	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseCommentDBConnection()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.WriteInternalError(&w, "Unexpected error")
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a valid json")
		return
	}

	var idResult string
	err = json.Unmarshal(objmap["idResult"], &idResult)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'content'")
		return
	}

	var tag string
	err = json.Unmarshal(objmap["tag"], &tag)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'content'")
		return
	}

	result := dbHandler.GetResultByID(idResult)
	if result == nil {
		utils.WriteInternalError(&w, dbError)
		return
	}

	result.Tags = helper.RemoveFromSlice(result.Tags, tag)

	ok := dbHandler.UpdateResult(result)
	if !ok {
		utils.WriteInternalError(&w, dbError)
		return
	}

	utils.ReturnSuccess(&w, "tag deleted successfully")
}
