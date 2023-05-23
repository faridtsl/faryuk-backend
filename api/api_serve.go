package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"FaRyuk/config"
	"FaRyuk/internal/types"
	"FaRyuk/internal/user"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	jwtCookieName   = "jwt-token"
	dbError         = "Database error"
	unexpectedError = "Unexpected error"
)

var backgroundScans = 0
var startTime time.Time

func initKeys() {
	APIRegisterKey = uuid.New().String()
	JWTSecret = uuid.New().String()

	fmt.Printf("REGISTER KEY %s\n", APIRegisterKey)
	fmt.Printf("JWT SECRET %s\n", JWTSecret)
}

func getCookie(name string, r *http.Request) (string, error) {
	tokenCookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	val := tokenCookie.Value
	return val, nil
}

func getIdentity(w *http.ResponseWriter, r *http.Request) (string, string, error) {
	token, err := getCookie(jwtCookieName, r)
	if err != nil {
		writeForbidden(w, "Cookie error")
		return "", "", err
	}

	username, idUser, err := user.GetUsername(token, JWTSecret)
	if err != nil {
		writeForbidden(w, "Authorization error")
		return "", "", err
	}
	return username, idUser, nil
}

func writeResponse(w *http.ResponseWriter, m types.JSONReturn) {
	(*w).Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(*w).Encode(m)
	if err != nil {
		return
	}
}

func writeForbidden(w *http.ResponseWriter, m string) {
	(*w).WriteHeader(http.StatusForbidden)
	writeResponse(w, types.JSONReturn{Status: "Fail", Body: m})
}

func writeNotFound(w *http.ResponseWriter, m string) {
	(*w).WriteHeader(http.StatusNotFound)
	writeResponse(w, types.JSONReturn{Status: "Fail", Body: m})
}

func writeInternalError(w *http.ResponseWriter, m string) {
	(*w).WriteHeader(http.StatusInternalServerError)
	writeResponse(w, types.JSONReturn{Status: "Fail", Body: m})
}

func writeObject(w *http.ResponseWriter, m interface{}) {
	writeResponse(w, types.JSONReturn{Status: "Success", Body: m})
}

// HandleRequests : set up routes for API
func HandleRequests() {
	initKeys()
	startTime = time.Now()
	myRouter := mux.NewRouter().StrictSlash(true)

	secure := myRouter.PathPrefix("/").Subrouter()
	secure.Use(verifyJWT)

	adminRouter := myRouter.PathPrefix("/").Subrouter()
	adminRouter.Use(verifyAdmin)

	// Users' endpoints
	addUserEndpoints(secure)

	// Groups' endpoints
	addGroupEndpoints(secure, adminRouter)

	// Result's endpoints
	addResultEndpoints(secure)

	// History endpoints
	addHistoryEndpoints(secure)

	// Sharing endpoints
	addSharingEndpoints(secure)

	// Comments endpoints
	addCommentEndpoints(secure)

	// Scans endpoints
	addScanEndpoints(secure)

	// Lists helper
	secure.HandleFunc("/api/get-dnslists", getDnsLists).Methods("GET")
	secure.HandleFunc("/api/get-wordlists", getWordLists).Methods("GET")
	secure.HandleFunc("/api/get-portlists", getPortLists).Methods("GET")

	// Scanners endpoints
	addRunnersEndpoints(secure)

	// App infos
	secure.HandleFunc("/api/infos", getInfos).Methods("GET")

	// Auth/Register endpoints
	myRouter.HandleFunc("/api/register", register).Methods("POST")
	myRouter.HandleFunc("/api/login", login).Methods("POST")
	secure.HandleFunc("/api/logout", logout).Methods("GET")

	listenAddr := fmt.Sprintf("%s:%d", config.Cfg.Server.Addr, config.Cfg.Server.Port)

	log.Fatal(http.ListenAndServe(listenAddr, myRouter))
}
