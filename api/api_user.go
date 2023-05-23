package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"FaRyuk/internal/db"
	"FaRyuk/internal/types"
	"FaRyuk/internal/user"

	"github.com/gorilla/mux"
)

// APIRegisterKey : This key should be used to register
var APIRegisterKey string

// JWTSecret : This key is used to sign JWT tokens
var JWTSecret string

const obfuscatedPassword = "*********"

func addUserEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/get-username/{id}", getUsername).Methods("GET")
	secure.HandleFunc("/api/get-users", getUsers).Methods("GET")
	secure.HandleFunc("/api/get-group-users", getUsersByGroup).Methods("POST")
	secure.HandleFunc("/api/whoami", whoami).Methods("GET")
	secure.HandleFunc("/api/isAdmin", isAdmin).Methods("GET")
	secure.HandleFunc("/api/change-password", changePassword).Methods("POST")
	secure.HandleFunc("/api/change-theme", changeTheme).Methods("POST")
}

func register(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	var key string
	err = json.Unmarshal(objmap["API_REGISTER_KEY"], &key)
	if err != nil || key != APIRegisterKey {
		writeForbidden(&w, "Please provide a valid API register key")
		return
	}

	var username string
	err = json.Unmarshal(objmap["username"], &username)
	if err != nil || username == "" {
		writeInternalError(&w, "Please provide a valid 'username'")
		return
	}

	var password string
	err = json.Unmarshal(objmap["password"], &password)
	if err != nil || password == "" {
		writeInternalError(&w, "Please provide a 'password'")
		return
	}

	var password2 string
	err = json.Unmarshal(objmap["password2"], &password2)
	if err != nil || password2 == "" || password != password2 {
		writeInternalError(&w, "Passwords don't match")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	usr := dbHandler.GetUserByUsername(username)

	if usr != nil {
		writeForbidden(&w, "User already exists")
		return
	}

	u := user.NewUser(username, password)
	err = dbHandler.InsertUser(u)

	if err != nil {
		writeInternalError(&w, dbError)
	}
	writeObject(&w, "User created succesfully")
}

func login(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	var username string
	err = json.Unmarshal(objmap["username"], &username)
	if err != nil || username == "" {
		writeInternalError(&w, "Please provide a valid 'username'")
		return
	}

	var password string
	err = json.Unmarshal(objmap["password"], &password)
	if err != nil || password == "" {
		writeInternalError(&w, "Please provide a 'password'")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	usr := dbHandler.GetUserByUsername(username)

	if usr == nil || !user.Login(usr, password) {
		writeForbidden(&w, "Wrong password or username")
		return
	}

	token, err := user.GenerateJWT(usr, JWTSecret)

	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: jwtCookieName, Value: token, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	writeObject(&w, token)
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	_, userID, err := getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Invalide token")
		return
	}

	var currentPassword string
	err = json.Unmarshal(objmap["currentPassword"], &currentPassword)
	if err != nil || currentPassword == "" {
		writeInternalError(&w, "Please provide a 'current password'")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	usr := dbHandler.GetUserByID(userID)
	if usr == nil || !user.Login(usr, currentPassword) {
		writeForbidden(&w, "Wrong password")
		return
	}

	var password string
	err = json.Unmarshal(objmap["password"], &password)
	if err != nil || password == "" {
		writeInternalError(&w, "Please provide a 'password'")
		return
	}

	var password2 string
	err = json.Unmarshal(objmap["password2"], &password2)
	if err != nil || password2 == "" || password != password2 {
		writeInternalError(&w, "Passwords don't match")
		return
	}

	usr.Password, err = user.GetHashedPassword(password)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = dbHandler.UpdateUser(usr)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	writeObject(&w, "Password changed successfully")
}

func changeTheme(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	_, userID, err := getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Invalide token")
		return
	}

	var theme string
	err = json.Unmarshal(objmap["theme"], &theme)
	if err != nil || theme == "" {
		writeInternalError(&w, "Please provide a 'theme'")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	usr := dbHandler.GetUserByID(userID)
	if usr == nil {
		writeForbidden(&w, "Wrong user")
		return
	}

	usr.Theme = theme
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = dbHandler.UpdateUser(usr)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	writeObject(&w, "Theme changed successfully")
}

func whoami(w http.ResponseWriter, r *http.Request) {
	_, id, err := getIdentity(&w, r)
	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	usr := dbHandler.GetUserByID(id)
	if err != nil {
		return
	}

	if usr.Theme == "" {
		usr.Theme = "light"
	}

	usr.Password = obfuscatedPassword

	writeObject(&w, usr)
}

func isAdmin(w http.ResponseWriter, r *http.Request) {
	username, _, err := getIdentity(&w, r)
	if err != nil {
		return
	}
	writeObject(&w, username == "admin")
}

func getUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idUser := vars["id"]

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	usr := dbHandler.GetUserByID(idUser)

	if usr == nil {
		writeForbidden(&w, "User not found")
		return
	}
	writeObject(&w, usr.Username)
}

func getUsersByGroup(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage
	var group types.Group

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	err = json.Unmarshal(body, &objmap)
	if err != nil {
		writeInternalError(&w, "Please provide a valid json")
		return
	}

	_, _, err = getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Invalide token")
		return
	}

	var groupid string
	err = json.Unmarshal(objmap["idGroup"], &groupid)
	if err != nil || groupid == "" {
		writeInternalError(&w, "Please provide a 'groupid'")
		return
	}

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	group, err = dbHandler.GetGroupByID(groupid)
	if err != nil {
		writeInternalError(&w, dbError)
		return
	}

	usrs, err := dbHandler.GetUsersByGroup(group)

	if err != nil {
		writeInternalError(&w, dbError)
		return
	}

	for idx := range usrs {
		usrs[idx].Password = "**********"
	}

	writeObject(&w, usrs)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()
	usrs := dbHandler.GetUsers()

	for idx := range usrs {
		usrs[idx].Password = "**********"
	}

	writeObject(&w, usrs)
}

func logout(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: jwtCookieName, Value: "", Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)
	writeObject(&w, "Logged off successfully")
}

func verifyJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(jwtCookieName)

		if err != nil {
			writeForbidden(&w, "Invalid token")
			return
		}

		token := tokenCookie.Value
		if !user.VerifyJWT(token, JWTSecret) {
			writeForbidden(&w, "Invalid token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func verifyAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(jwtCookieName)

		if err != nil {
			writeForbidden(&w, "Invalid token")
			return
		}

		token := tokenCookie.Value
		if !user.VerifyJWT(token, JWTSecret) {
			writeForbidden(&w, "Invalid token")
			return
		}
		username, _, err := user.GetUsername(token, JWTSecret)
		if err != nil || username != "admin" {
			writeForbidden(&w, "Invalid token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
