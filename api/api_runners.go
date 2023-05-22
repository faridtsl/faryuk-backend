package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"FaRyuk/api/utils"
	"FaRyuk/internal/runner"
	"FaRyuk/internal/types"
	"FaRyuk/models"

	"github.com/gorilla/mux"
)

func addRunnersEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/scanners", getRunners).Methods("GET")
	secure.HandleFunc("/api/scanner", addRunner).Methods("POST")
	secure.HandleFunc("/api/scanner", deleteRunner).Methods("DELETE")
}

func getRunners(w http.ResponseWriter, r *http.Request) {
	var err error
	var runners []types.Runner
	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()

	username, idUser, err := getIdentity(&w, r)
	if err != nil {
		return
	}

	if username == "admin" {
		runners, err = dbHandler.GetRunners()

		if err != nil {
			utils.WriteInternalError(&w, "Database error")
			return
		}
	} else {
		runners, err = dbHandler.GetRunnersByUserID(idUser)
		if err != nil {
			utils.WriteInternalError(&w, "Database error")
			return
		}
	}

	utils.ReturnSuccess(&w, runners)
}

func addRunner(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage
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

	_, idUser, err := getIdentity(&w, r)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a valid identity")
		return
	}

	var tag string
	err = json.Unmarshal(objmap["tag"], &tag)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'tag'")
		return
	}

	var displayName string
	err = json.Unmarshal(objmap["displayName"], &displayName)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'displayName'")
		return
	}

	var cmdLine string
	err = json.Unmarshal(objmap["cmd"], &cmdLine)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'cmd'")
		return
	}

	var isWeb bool
	err = json.Unmarshal(objmap["isWeb"], &isWeb)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'isWeb'")
		return
	}

	var isPort bool
	err = json.Unmarshal(objmap["isPort"], &isPort)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'isPort'")
		return
	}

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()

	runner := runner.NewRunner(tag, displayName, strings.Split(cmdLine, " "), idUser, isWeb, isPort)
	err = dbHandler.InsertRunner(runner)
	if err != nil {
		utils.WriteInternalError(&w, "Database error")
		return
	}
	utils.ReturnSuccess(&w, "Runner added")
}

func deleteRunner(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage
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

	var id string
	err = json.Unmarshal(objmap["id"], &id)
	if err != nil {
		utils.WriteInternalError(&w, "Please provide a 'id'")
		return
	}

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()

	err = dbHandler.RemoveRunnerByID(id)
	if err != nil {
		utils.WriteInternalError(&w, "Database error")
		return
	}
	utils.ReturnSuccess(&w, "Runner deleted")
}
