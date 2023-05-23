package api

import (
	"net/http"
	"strconv"

	"FaRyuk/internal/db"
	"FaRyuk/internal/group"
	"FaRyuk/internal/helper"
	"FaRyuk/internal/types"

	"github.com/gorilla/mux"
)

func addHistoryEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/get-history", getHistory).Methods("GET")
	secure.HandleFunc("/api/count-history", countHistory).Methods("GET")
	secure.HandleFunc("/api/history/{id}", getHistoryRecordByID).Methods("GET")
	secure.HandleFunc("/api/history/{id}", deleteHistory).Methods("DELETE")
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	var results []types.HistoryRecord
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

	pageSizeSlice, ok := query["size"]
	pageSize := 10
	if ok {
		pageSize, _ = strconv.Atoi(pageSizeSlice[0])
	}

	offsetSlice, ok := query["offset"]
	offset := 1
	if ok {
		offset, _ = strconv.Atoi(offsetSlice[0])
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	if searchMap["group"] != "" {
		group, err := dbHandler.GetGroupsByName(searchMap["group"])
		if err != nil {
			writeInternalError(&w, dbError)
			return
		}
		searchMap["group"] = group.ID
	}

	if username == "admin" {
		results, err = dbHandler.GetHistoryRecordsBySearch(searchMap, offset, pageSize)
	} else {
		user := dbHandler.GetUserByID(idUser)
		results, err = dbHandler.GetHistoryRecordsBySearchAndOwner(searchMap, idUser, group.ToIDsArray(user.Groups), offset, pageSize)
	}
	if err != nil {
		writeInternalError(&w, "Cannot retrieve history")
		return
	}
	if len(results) == 0 {
		writeObject(&w, results)
		return
	}

	writeObject(&w, results)
}

func countHistory(w http.ResponseWriter, r *http.Request) {
	var cntRecords int
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

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	if username == "admin" {
		cntRecords, err = dbHandler.CountHistoryRecordsBySearch(searchMap)
	} else {
		user := dbHandler.GetUserByID(idUser)
		cntRecords, err = dbHandler.CountHistoryRecordsBySearchAndOwner(searchMap, group.ToIDsArray(user.Groups), idUser)
	}
	if err != nil {
		writeInternalError(&w, "Cannot retrieve history count")
		return
	}
	writeObject(&w, cntRecords)
}

func getHistoryRecordByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	result, err := dbHandler.GetHistoryRecordByID(id)
	if err != nil {
		writeInternalError(&w, "Cannot retrieve history record")
		return
	}
	writeObject(&w, result)
}

func deleteHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]
	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	res := dbHandler.RemoveHistoryRecordByID(id)
	if !res {
		writeInternalError(&w, "Cannot delete history record")
		return
	}
	writeObject(&w, "History record deleted successfully")
}
