package operations

import (
	"fmt"
	"sync"
	"time"

	"FaRyuk/internal/helper"
	"FaRyuk/internal/types"
	"FaRyuk/models"
	"FaRyuk/pkg"

	"github.com/google/uuid"
)

// DoHost : launchs scan on host
func DoHost(
	idUser string,
	host string,
	groupId string,
	portsFilename string,
	dirsFilename string,
	scanners []string,
) (types.Result, error) {
	var result types.Result
	var runners []types.Runner
	var webRunners []types.Runner
	ports := helper.FileToInts("./ressources/ports/" + portsFilename)
	dirs := helper.FileToStrings("./ressources/dirs/" + dirsFilename)

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()

	for idx := range scanners {
		r, err := dbHandler.GetRunnerByID(scanners[idx])
		if err != nil {
			continue
		}
		if r.IsWeb {
			webRunners = append(webRunners, r)
		} else {
			runners = append(runners, r)
		}
	}

	historyRecord := types.HistoryRecord{
		ID:          uuid.New().String(),
		Owner:       idUser,
		IsWeb:       false,
		IsFinished:  false,
		Host:        host,
		OwnerGroup:  groupId,
		CreatedDate: time.Now(),
	}
	historyRecord.State = append(historyRecord.State, "[*] Scan started", "[*] Portlist : "+portsFilename, "[*] Wordlist : "+dirsFilename)
	err := dbHandler.InsertHistoryRecord(historyRecord)
	if err != nil {
		return result, fmt.Errorf("could not create history record")
	}

	historyUpdater := func(statement string) {
		historyRecord.State = append(historyRecord.State, statement)
		dbHandler.UpdateHistoryRecord(historyRecord)
	}
	// Resolve domain
	resolver := pkg.NewResolver()
	resolutions := resolver.Resolve(host)

	if len(resolutions) == 0 {
		historyUpdater("[-] Resolution failed")
		historyRecord.IsFinished = true
		historyRecord.IsSuccess = false
		historyUpdater("[-] Scan failed")
		return result, fmt.Errorf("resolutions failed")
	}

	historyUpdater("[+] Resolution finished")

	// Scan ports
	portscanner := pkg.NewPortScanner(host, 2*time.Second, 5)
	openPorts := portscanner.Run(ports)
	historyUpdater("[+] Port scanning finished : " + fmt.Sprintf("%v", openPorts))

	result = types.Result{
		ID:          uuid.New().String(),
		Owner:       idUser,
		Host:        host,
		Ips:         resolutions,
		OpenPorts:   openPorts,
		OwnerGroup:  groupId,
		CreatedDate: time.Now(),
	}

	for idx := range runners {
		r, err := launchRunner(host, 0, "", runners[idx])
		if err != nil {
			result.Err = append(result.Err, fmt.Sprintf("%s", err))
		} else {
			result.RunnerOutput = append(result.RunnerOutput, r)
		}
	}

	for _, port := range result.OpenPorts {
		isWeb, isSSL := fingerprintPort(host, port)
		if isWeb {
			webresult, _ := getWebResult(idUser, host, port, isSSL, "", dirs, "", false, "", historyRecord.ID, webRunners)
			result.WebResults = append(result.WebResults, webresult)
		}
	}

	if !helper.ContainsStr(result.Tags, "#new") {
		result.Tags = append(result.Tags, "#new")
	}

	historyRecord, _ = dbHandler.GetHistoryRecordByID(historyRecord.ID)
	historyRecord.IsFinished = true
	historyRecord.IsSuccess = true
	historyUpdater("[+] Scan finished")
	return result, nil
}

// WebScanPort : launches a webscan of a host in a given port
func WebScanPort(idUser, id string, port int, ssl bool, base, dirFilename, statusCodes string, wildcardForced bool, excludedText string, scanners []string) bool {
	var res types.Result
	var webRunners []types.Runner
	dirs := helper.FileToStrings("./ressources/dirs/" + dirFilename)
	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()

	for idx := range scanners {
		r, err := dbHandler.GetRunnerByID(scanners[idx])
		if err != nil {
			continue
		}
		if r.IsWeb {
			webRunners = append(webRunners, r)
		}
	}

	resPtr := dbHandler.GetResultByID(id)
	if resPtr == nil {
		return false
	}
	res = *resPtr
	res.Owner = idUser
	webresult, _ := getWebResult(idUser, res.Host, port, ssl, base, dirs, statusCodes, wildcardForced, excludedText, "-"+dirFilename, webRunners)
	webresult.CreatedDate = time.Now()
	l := sync.Mutex{}
	l.Lock()
	resPtr = dbHandler.GetResultByID(id)
	res = *resPtr
	var exists = false
	for idx, webres := range res.WebResults {
		if webres.Port != port {
			continue
		}
		exists = true
		set := make(map[string]bool)
		for _, busterres := range webres.Busterres {
			set[busterres.Path] = true
		}

		for _, busterres := range webresult.Busterres {
			_, present := set[busterres.Path]
			if !present {
				res.WebResults[idx].Busterres = append(res.WebResults[idx].Busterres, busterres)
			}
		}
		res.WebResults[idx].Err = append(res.WebResults[idx].Err, webresult.Err...)
		res.WebResults[idx].RunnerOutput = webresult.RunnerOutput
	}
	if !exists {
		res.WebResults = append(res.WebResults, webresult)
	}

	if !helper.ContainsStr(res.Tags, "#new") {
		res.Tags = append(res.Tags, "#new")
	}

	retval := dbHandler.UpdateResult(&res)
	l.Unlock()
	return retval
}

func RunnerScanPort(idUser, id string, port int, scanners []string) bool {
	var portRunners []types.Runner
	var historyRecord types.HistoryRecord
	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()
	resPtr := dbHandler.GetResultByID(id)
	result := *resPtr

	historyRecord.ID = uuid.New().String()
	historyRecord.Owner = idUser
	historyRecord.IsWeb = false
	historyRecord.IsFinished = false
	historyRecord.Host = result.Host
	historyRecord.OwnerGroup = result.OwnerGroup
	historyRecord.State = append(historyRecord.State, "[*] Scan started")
	historyRecord.CreatedDate = time.Now()
	err := dbHandler.InsertHistoryRecord(historyRecord)
	if err != nil {
		return false
	}

	for idx := range scanners {
		r, err := dbHandler.GetRunnerByID(scanners[idx])
		if err != nil {
			continue
		}
		if r.IsPort && !r.IsWeb {
			portRunners = append(portRunners, r)
		}
	}

	for idx := range portRunners {
		historyRecord.State = append(historyRecord.State,
			"[*] "+portRunners[idx].DisplayName+" started for port "+fmt.Sprintf("%d", port))
		dbHandler.UpdateHistoryRecord(historyRecord)
		r, err := launchRunner(result.Host, port, "", portRunners[idx])
		if err == nil {
			exists := false
			for idx := range result.RunnerOutput {
				if result.RunnerOutput[idx].ToolName == r.ToolName && result.RunnerOutput[idx].ScannedPort == fmt.Sprintf("%d", port) {
					result.RunnerOutput[idx] = r
					exists = true
				}
			}
			if !exists {
				result.RunnerOutput = append(result.RunnerOutput, r)
			}
			historyRecord.State = append(historyRecord.State,
				"[+] "+portRunners[idx].DisplayName+" finished for port "+fmt.Sprintf("%d", port))
		} else {
			historyRecord.State = append(historyRecord.State,
				"[-] "+portRunners[idx].DisplayName+" failed for port "+fmt.Sprintf("%d", port))
			result.Err = append(result.Err, fmt.Sprintf("%s", err))
		}
		dbHandler.UpdateHistoryRecord(historyRecord)
	}

	if !helper.ContainsStr(result.Tags, "#new") {
		result.Tags = append(result.Tags, "#new")
	}
	retval := dbHandler.UpdateResult(&result)

	historyRecord.IsFinished = true
	historyRecord.IsSuccess = true
	historyRecord.State = append(historyRecord.State, "[+] Scan finished")
	dbHandler.UpdateHistoryRecord(historyRecord)

	return retval
}

// DoDomain : launch gobuster on domain
func DoDomain(idUser, domain, groupId, subdomainFilename string, isWildcard bool, resolver string) ([]string, error) {
	var historyRecord types.HistoryRecord
	dirs := helper.FileToStrings("./ressources/subdomains/" + subdomainFilename)
	chunks := helper.ChunkSlice(dirs, len(dirs)/9)
	results := make([]string, 0)

	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseConnection()

	historyRecord.ID = uuid.New().String()
	historyRecord.Owner = idUser
	historyRecord.OwnerGroup = groupId
	historyRecord.IsWeb = false
	historyRecord.IsFinished = false
	historyRecord.Domain = domain
	historyRecord.State = append(historyRecord.State, "[*] Scan started", "[*] DNS list : "+subdomainFilename)
	historyRecord.CreatedDate = time.Now()
	err := dbHandler.InsertHistoryRecord(historyRecord)
	if err != nil {
		return results, fmt.Errorf("could not create history record")
	}

	for idx, slice := range chunks {
		r, err := launchBusterDNS(domain, slice, isWildcard, resolver)
		if err != nil {
			historyRecord.State = append(historyRecord.State, "[-] Scan failed : "+fmt.Sprintf("%s", err))
			historyRecord.IsFinished = true
			historyRecord.IsSuccess = false
			dbHandler.UpdateHistoryRecord(historyRecord)
			return make([]string, 0), err
		}
		results = append(results, r...)
		historyRecord.State = append(historyRecord.State, fmt.Sprintf("[*] : %%%d done, found : %d", (idx+1)*10, len(r)))
		dbHandler.UpdateHistoryRecord(historyRecord)
	}

	historyRecord.State = append(historyRecord.State, fmt.Sprintf("[+] Scan finished : Found %d", len(results)))
	historyRecord.IsFinished = true
	historyRecord.IsSuccess = true
	dbHandler.UpdateHistoryRecord(historyRecord)
	return results, nil
}
