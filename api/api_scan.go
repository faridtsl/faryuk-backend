package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"FaRyuk/internal/db"
	"FaRyuk/internal/helper"
	"FaRyuk/internal/operations"

	"github.com/gorilla/mux"
)

func addScanEndpoints(secure *mux.Router) {
	secure.HandleFunc("/api/scan", doScan).Methods("POST")
	secure.HandleFunc("/api/scan-multiple", doMultipleScan).Methods("POST")
	secure.HandleFunc("/api/domain-scan", doDomainScan).Methods("POST")
	secure.HandleFunc("/api/webscan", webscanResultByID).Methods("POST")
	secure.HandleFunc("/api/portscan", doPortScan).Methods("POST")
}

func webscanResultByID(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	unmarshal := func(src []byte, dest any, errStr string) error {
		err := json.Unmarshal(src, dest)
		if err != nil {
			writeInternalError(&w, errStr)
			return err
		}
		return nil
	}

	if unmarshal(body, &objmap, "Please provide a valid json") != nil {
		return
	}

	var id string
	if unmarshal(objmap["id"], &id, "Please provide a valid id") != nil {
		return
	}

	var base string
	if unmarshal(objmap["base"], &base, "Please provide a valid base") != nil {
		return
	}

	var statusCodes string
	if unmarshal(objmap["statusCodes"], &statusCodes, "Please provide a valid statusCodes") != nil {
		return
	}

	var ssl bool
	if unmarshal(objmap["ssl"], &ssl, "Please provide a valid ssl") != nil {
		return
	}

	var wildcardForced bool
	if unmarshal(objmap["useWildcard"], &wildcardForced, "Please provide a valid useWildcard option") != nil {
		return
	}

	var excludedText string
	if unmarshal(objmap["excludeBuster"], &excludedText, "Please provide a valid excludedText") != nil {
		return
	}

	var webPortstr string
	var webPort int
	err = json.Unmarshal(objmap["webPort"], &webPortstr)
	if err == nil {
		webPort, err = strconv.Atoi(webPortstr)
	}

	if err != nil {
		writeInternalError(&w, "Please provide a valid webPort")
		return
	}

	var wordlist string
	if unmarshal(objmap["wordlist"], &wordlist, "Please provide a valid wordlist") != nil {
		return
	}

	var scanners []string
	if unmarshal(objmap["scanners"], &scanners, "Please provide a valid scanners") != nil {
		return
	}

	_, idUser, err := getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Identity error")
		return
	}

	go operations.WebScanPort(idUser,
		id,
		webPort,
		ssl,
		base,
		wordlist,
		statusCodes,
		wildcardForced,
		excludedText,
		scanners)

	writeObject(&w, "Webscan started")
}

func scanAndSave(idUser string, host string, groupID string, portlist string, dirlist string, rescan bool, scanners []string) {

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	rs, _ := dbHandler.GetResultsByHostAndOwner(host, idUser)
	l := sync.Mutex{}
	if len(rs) > 0 {
		if !rescan {
			backgroundScans--
			return
		}
		result, err := operations.DoHost(idUser, host, groupID, portlist, dirlist, scanners)
		if err != nil {
			backgroundScans--
			return
		}
		l.Lock()
		orig := dbHandler.GetResultByID(rs[0].ID)
		for _, wr := range result.WebResults {
			// Check if port is already in original result
			exists := false
			idxOrig := -1
			for idx, wrOrig := range orig.WebResults {
				if wrOrig.Port == wr.Port {
					// Port found
					exists = true
					idxOrig = idx
					break
				}
			}

			if !exists {
				// Add newly scanned port
				orig.WebResults = append(orig.WebResults, wr)
				continue
			}

			// Merge web results
			orig.WebResults[idxOrig].Screen = wr.Screen
			for _, busterRes := range wr.Busterres {
				exists = false
				// Check if dir is already found
				for _, busterOrig := range orig.WebResults[idxOrig].Busterres {
					if busterOrig.Path == busterRes.Path {
						exists = true
						break
					}
				}
				if !exists {
					orig.WebResults[idxOrig].Busterres = append(orig.WebResults[idxOrig].Busterres, busterRes)
				}
			}
		}

		if rescan {
			for _, port := range result.OpenPorts {
				if !helper.Contains(orig.OpenPorts, port) {
					orig.OpenPorts = append(orig.OpenPorts, port)
				}
			}
		}

		if !helper.ContainsStr(orig.Tags, "#new") {
			orig.Tags = append(orig.Tags, "#new")
		}
		orig.OwnerGroup = groupID
		dbHandler.UpdateResult(orig)
		l.Unlock()
	} else {
		result, err := operations.DoHost(idUser, host, groupID, portlist, dirlist, scanners)
		if err != nil {
			fmt.Println(err)
			return
		}
		l.Lock()
		err = dbHandler.InsertResult(&result)
		l.Unlock()
		if err != nil {
			backgroundScans--
			return
		}
	}
	backgroundScans--
}

func doScan(w http.ResponseWriter, r *http.Request) {
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

	var host string
	err = json.Unmarshal(objmap["host"], &host)
	if err != nil || host == "" {
		writeInternalError(&w, "Please provide a valid host")
		return
	}

	var groupID string
	err = json.Unmarshal(objmap["idGroup"], &groupID)
	if err != nil {
		return
	}

	var dirlistFilename string
	err = json.Unmarshal(objmap["dirlist"], &dirlistFilename)
	if err != nil || dirlistFilename == "" {
		writeInternalError(&w, "Please provide a valid dirlist")
		return
	}

	var portlistFilename string
	err = json.Unmarshal(objmap["portlist"], &portlistFilename)
	if err != nil || portlistFilename == "" {
		writeInternalError(&w, "Please provide a valid portlist")
		return
	}

	var rescan bool
	err = json.Unmarshal(objmap["rescan"], &rescan)
	if err != nil {
		writeInternalError(&w, "Please provide a valid rescan option")
		return
	}

	var scanners []string
	err = json.Unmarshal(objmap["scanners"], &scanners)
	if err != nil && objmap["scanners"] != nil {
		writeInternalError(&w, err.Error())
		return
	}

	_, idUser, err := getIdentity(&w, r)
	if err != nil {
		return
	}

	backgroundScans++
	go scanAndSave(idUser, html.EscapeString(host), groupID, portlistFilename, dirlistFilename, rescan, scanners)
	writeObject(&w, "Scan started")
}

func doMultipleScan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeInternalError(&w, "Form parsing error")
		return
	}
	_, idUser, err := getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Identity error")
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		writeInternalError(&w, fmt.Sprintf("%s", err))
		return
	}
	file, _, err := r.FormFile("hosts")
	if err != nil {
		writeInternalError(&w, fmt.Sprintf("%s", err))
		return
	}
	defer file.Close()

	hosts := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hosts = append(hosts, html.EscapeString(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		writeInternalError(&w, "Internal error")
		return
	}

	groupID := r.PostForm["idGroup"][0]
	rescan := true
	if len(r.PostForm["rescan"]) == 0 {
		rescan = false
	}
	dirlistFilename := r.PostForm["dirlist"][0]
	portlistFilename := r.PostForm["portlist"][0]
	scanners := r.PostForm["scanners"]
	go scanMultipleAndSave(idUser, hosts, groupID, portlistFilename, dirlistFilename, rescan, scanners)
	writeObject(&w, "Scan multiple started")
}

func scanMultipleAndSave(idUser string, hosts []string,
	groupID string,
	portlist string, dirlist string,
	rescan bool, scanners []string) {
	backgroundScans += len(hosts)
	sem := make(chan bool, 5)
	for _, host := range hosts {
		sem <- true
		go func(host string) {
			scanAndSave(idUser, host, groupID, portlist, dirlist, rescan, scanners)
			<-sem
		}(host)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func scanDomainAndSave(idUser string, domain string,
	groupID string,
	dnslistFilename string, portlistFilename string,
	dirlistFilename string, resolver string,
	isWildcard bool, rescan bool, scanners []string) {

	hosts, _ := operations.DoDomain(idUser, domain, groupID, dnslistFilename, isWildcard, resolver)
	for idx := range hosts {
		hosts[idx] = hosts[idx] + "." + domain
	}
	go scanMultipleAndSave(idUser, hosts, groupID, portlistFilename, dirlistFilename, rescan, scanners)
}

func doPortScan(w http.ResponseWriter, r *http.Request) {
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

	var id string
	err = json.Unmarshal(objmap["id"], &id)
	if err != nil {
		writeInternalError(&w, "Please provide a valid id")
		return
	}

	var scannedPort string
	err = json.Unmarshal(objmap["scannedPort"], &scannedPort)
	if err != nil {
		fmt.Println(err)
		writeInternalError(&w, "Please provide a valid scannedPort")
		return
	}

	port, err := strconv.Atoi(scannedPort)
	if err != nil {
		fmt.Println(err)
		writeInternalError(&w, "Please provide a valid port")
		return
	}

	var scanners []string
	err = json.Unmarshal(objmap["scanners"], &scanners)
	if err != nil {
		writeInternalError(&w, err.Error())
		return
	}

	_, idUser, err := getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Identity error")
		return
	}

	go operations.RunnerScanPort(idUser,
		id,
		port,
		scanners)

	writeObject(&w, "port runner scan started")
}

func doDomainScan(w http.ResponseWriter, r *http.Request) {
	var objmap map[string]json.RawMessage

	_, idUser, err := getIdentity(&w, r)
	if err != nil {
		writeInternalError(&w, "Identity error")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeInternalError(&w, unexpectedError)
		return
	}

	unmarshal := func(src []byte, dest any, errStr string) error {
		err := json.Unmarshal(src, dest)
		if err != nil {
			writeInternalError(&w, errStr)
			return err
		}
		return nil
	}

	if unmarshal(body, &objmap, "Please provide a valid json") != nil {
		return
	}

	var domain string
	if unmarshal(objmap["domain"], &domain, "Please provide a valid domain") != nil {
		return
	}

	var groupID string
	if unmarshal(objmap["idGroup"], &groupID, "Please provide a valid idGroup") != nil {
		return
	}

	var dirlistFilename string
	if unmarshal(objmap["dirlist"], &dirlistFilename, "Please provide a valid dirlist") != nil || dirlistFilename == "" {
		writeInternalError(&w, "Please provide a valid dirlist")
		return
	}

	var portlistFilename string
	if unmarshal(objmap["portlist"], &portlistFilename, "Please provide a valid portlist") != nil || portlistFilename == "" {
		writeInternalError(&w, "Please provide a valid portlist")
		return
	}

	var dnslistFilename string
	if unmarshal(objmap["dnslist"], &dnslistFilename, "Please provide a valid dnslist") != nil || dnslistFilename == "" {
		writeInternalError(&w, "Please provide a valid dnslist")
		return
	}

	var resolver string
	if unmarshal(objmap["resolver"], &resolver, "Please provide a valid resolver") != nil {
		return
	}

	var rescan bool
	if unmarshal(objmap["rescan"], &rescan, "Please provide a valid rescan") != nil {
		return
	}

	var wildcard bool
	if unmarshal(objmap["wildcard"], &wildcard, "Please provide a valid wildcard") != nil {
		return
	}

	var scanners []string
	if unmarshal(objmap["scanners"], &scanners, "Please provide a valid scanners") != nil {
		return
	}

	go scanDomainAndSave(idUser, domain, groupID,
		dnslistFilename, portlistFilename,
		dirlistFilename, resolver,
		wildcard, rescan, scanners)

	writeObject(&w, "Scan domain started")
}

func getDnsLists(w http.ResponseWriter, r *http.Request) {
	var dnsList = helper.GetDNSlists()
	writeObject(&w, dnsList)
}

func getPortLists(w http.ResponseWriter, r *http.Request) {
	var portlist = helper.GetPortlists()
	writeObject(&w, portlist)
}

func getWordLists(w http.ResponseWriter, r *http.Request) {
	var wordlist = helper.GetWordlists()
	writeObject(&w, wordlist)
}
