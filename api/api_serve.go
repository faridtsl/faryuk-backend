package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"FaRyuk/api/comment"
	"FaRyuk/api/utils"
	"FaRyuk/config"
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
		utils.WriteForbidden(w, "Cookie error")
		return "", "", err
	}

	username, idUser, err := user.GetUsername(token, JWTSecret)
	if err != nil {
		utils.WriteForbidden(w, "Authorization error")
		return "", "", err
	}
	return username, idUser, nil
}

// HandleRequests : set up routes for API
func HandleRequests() {
	initKeys()
	fmt.Println(" loading config")
	cfg, _ := config.MakeConfig()
	fmt.Println("config loaded", cfg)

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
	comment.AddCommentEndpoints(secure)

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

	listenAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Println("running on", cfg.Server.Host, ":", cfg.Server.Port)

	log.Fatal(http.ListenAndServe(listenAddr, myRouter))
}
