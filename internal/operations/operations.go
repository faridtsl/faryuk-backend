package operations

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"FaRyuk/internal/db"
	"FaRyuk/internal/runner"
	"FaRyuk/internal/types"
	"FaRyuk/pkg"

	"github.com/google/uuid"
)

func launchBusterDNS(
	domain string,
	dirs []string,
	wildCardForced bool,
	resolver string,
) ([]string, error) {
	// Scan subdomains
	opts := pkg.NewOptionsDNS(domain, wildCardForced, resolver)

	opts.Timeout = 1 * time.Second

	busterdns, err := pkg.NewGobusterDNS(opts)
	if err != nil {
		return nil, err
	}

	err = busterdns.PreRun()
	if err != nil {
		return nil, err
	}

	return busterdns.Run(dirs), nil
}

func launchBuster(
	url string,
	dirs []string,
	sCodes string,
	wildCardForced bool,
	excludedText string,
) ([]pkg.GoBusterResult, error) {
	// Scan dirs
	headers := []string{}
	statusCodes := "200,204,301,302,307,401,403"

	if sCodes != "" {
		statusCodes = sCodes
	}

	opts := pkg.NewOptionsDir(statusCodes, headers, wildCardForced, excludedText)

	opts.URL = url
	opts.Cookies = ""
	opts.FollowRedirect = false
	opts.NoTLSValidation = true
	opts.Timeout = 10 * time.Second
	opts.Username = ""
	opts.Password = ""
	opts.UserAgent = "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0"

	ctx := context.Background()

	buster, err := pkg.NewGobusterDir(ctx, opts)
	if err != nil {
		return nil, err
	}

	err = buster.PreRun()
	if err != nil {
		return nil, err
	}

	return buster.Run(dirs), nil
}

func getWebResult(
	idUser string,
	host string,
	port int,
	ssl bool,
	base string,
	dirs []string,
	statusCodes string,
	wildcardForced bool,
	excludedText string,
	historyID string,
	runners []types.Runner,
) (types.WebResult, error) {
	var err error
	var url string
	var historyRecord types.HistoryRecord

	dbHandler := db.NewDBHandler()
	defer dbHandler.CloseConnection()

	webresult := types.WebResult{}

	if historyID[0] == '-' {
		historyRecord = types.HistoryRecord{
			ID:          uuid.New().String(),
			Owner:       idUser,
			IsWeb:       true,
			IsFinished:  false,
			Host:        host,
			CreatedDate: time.Now(),
		}
		historyRecord.State = append(historyRecord.State, "[*] Scan started")
		historyRecord.State = append(historyRecord.State, "[*] Port : "+fmt.Sprintf("%d", port))
		historyRecord.State = append(historyRecord.State, "[*] Base : "+base)
		historyRecord.State = append(historyRecord.State, "[*] Wordlist : "+historyID[1:])
		err = dbHandler.InsertHistoryRecord(historyRecord)
	} else {
		historyRecord, err = dbHandler.GetHistoryRecordByID(historyID)
	}
	if err != nil {
		return webresult, err
	}

	webresult.Port = port
	webresult.Ssl = ssl

	proto := "http://"
	if ssl {
		proto = "https://"
	}

	url = fmt.Sprintf("%s%s:%d", proto, host, port)

	// Headergrab
	p := pkg.NewHeaderGrabber()
	webresult.Headers, err = p.Run(url)
	if err != nil {
		webresult.Err = append(webresult.Err, fmt.Sprintf("%s", err))
		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[-] Headers grabbing failed for port %d", port),
		)
	} else {
		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[+] Headers grabbed for port %d", port),
		)
	}
	dbHandler.UpdateHistoryRecord(historyRecord)

	// Screen homepage
	screener := pkg.NewScreener()
	webresult.Screen, err = screener.Run(url)
	if err != nil {
		webresult.Err = append(webresult.Err, fmt.Sprintf("%s", err))
		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[-] Screenshot failed for port %d", port),
		)
	} else {
		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[+] Screenshot for port %d", port),
		)
	}
	dbHandler.UpdateHistoryRecord(historyRecord)

	// GoBuster
	if base != "" {
		for idx := range dirs {
			dirs[idx] = base + "/" + dirs[idx]
		}
	}
	webresult.Busterres, err = launchBuster(url, dirs, statusCodes, wildcardForced, excludedText)
	if err != nil {
		webresult.Err = append(webresult.Err, fmt.Sprintf("%s", err))
		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[-] GoBuster failed for port %d / Error : %s", port, err),
		)
	} else {
		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[+] GoBuster finished for port %d / Found : %d", port, len(webresult.Busterres)),
		)
	}
	dbHandler.UpdateHistoryRecord(historyRecord)

	for idx := range runners {

		historyRecord.State = append(
			historyRecord.State,
			fmt.Sprintf("[*] %s started for port %d", runners[idx].DisplayName, port),
		)
		dbHandler.UpdateHistoryRecord(historyRecord)
		r, err := launchRunner(host, port, proto, runners[idx])
		if err == nil {
			exists := false
			for idx := range webresult.RunnerOutput {
				if webresult.RunnerOutput[idx].ToolName == r.ToolName {
					webresult.RunnerOutput[idx] = r
					exists = true
				}
			}
			if !exists {
				webresult.RunnerOutput = append(webresult.RunnerOutput, r)
			}
			historyRecord.State = append(
				historyRecord.State,
				fmt.Sprintf("[+] %s Finished for port %d", runners[idx].DisplayName, port),
			)
		} else {
			webresult.Err = append(webresult.Err, fmt.Sprintf("%s", err))
			historyRecord.State = append(
				historyRecord.State,
				fmt.Sprintf("[-] %s Failed for port %d", runners[idx].DisplayName, port),
			)
		}
		dbHandler.UpdateHistoryRecord(historyRecord)
	}

	if historyRecord.IsWeb {
		historyRecord.IsFinished = true
		historyRecord.IsSuccess = true
		dbHandler.UpdateHistoryRecord(historyRecord)
	}
	return webresult, nil
}

func fingerprintPort(host string, port int) (bool, bool) {
	if port == 80 {
		return true, false
	}
	if port == 443 {
		return true, true
	}

	url := fmt.Sprintf("http://%s:%d", host, port)
	_, err := http.Get(url)
	if err == nil {
		return true, false
	}

	url = fmt.Sprintf("https://%s:%d", host, port)
	_, err = http.Get(url)
	if err == nil {
		return true, true
	}

	return false, false
}

func launchRunner(host string, port int, proto string, r types.Runner) (types.RunnerResult, error) {
	p := fmt.Sprintf("%d", port)
	for idx := range r.Cmd {
		r.Cmd[idx] = strings.ReplaceAll(r.Cmd[idx], "[[host]]", host)
		r.Cmd[idx] = strings.ReplaceAll(r.Cmd[idx], "[[port]]", p)
		r.Cmd[idx] = strings.ReplaceAll(r.Cmd[idx], "[[proto]]", proto)
	}

	runnerHandler := runner.NewRunnerHandler()
	_, err := runnerHandler.PullImage(r.Tag)
	if err != nil {
		return types.RunnerResult{}, err
	}
	stdout, stderr, err := runnerHandler.RunCmd(r.Tag, r.Cmd)
	if err != nil {
		return types.RunnerResult{}, err
	}

	res := types.RunnerResult{}
	res.ID = uuid.New().String()
	res.Output = stdout
	res.Stderr = stderr
	res.ToolName = r.DisplayName
	res.ScannedPort = p

	return res, nil
}
