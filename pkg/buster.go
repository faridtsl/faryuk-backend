package pkg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/google/uuid"
)

// GoBusterResult a structure that stores results from GoBuster
type GoBusterResult struct {
	Path       string `bson:"path" json:"path"`
	StatusCode int    `bson:"statusCode" json:"statusCode"`
	Size       int    `bson:"size" json:"size"`
}

// OptionsDir is the struct to hold all options for this plugin
type OptionsDir struct {
	libgobuster.HTTPOptions
	StatusCodes       string
	StatusCodesParsed libgobuster.IntSet
	WildcardForced    bool
	ExcludeText       string
}

// GobusterDir is the main type to implement the interface
type GobusterDir struct {
	options    *OptionsDir
	globalopts *libgobuster.Options
	http       *libgobuster.HTTPClient
}

// NewGoBusterResult returns a new GoBusterResult struct
func NewGoBusterResult(path string, statusCode, size int) *GoBusterResult {
	return &GoBusterResult{path, statusCode, size}
}

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsDir(statusCodes string, headers []string, wildcard bool, excludeText string) *OptionsDir {

	var optheaders []libgobuster.HTTPHeader
	for _, h := range headers {
		keyAndValue := strings.SplitN(h, ":", 2)
		if len(keyAndValue) != 2 {
			continue
		}
		key := strings.TrimSpace(keyAndValue[0])
		value := strings.TrimSpace(keyAndValue[1])
		if key == "" {
			continue
		}
		header := libgobuster.HTTPHeader{Name: key, Value: value}
		optheaders = append(optheaders, header)
	}
	parsed, _ := helper.ParseCommaSeparatedInt(statusCodes)
	ret := &OptionsDir{
		StatusCodes:       statusCodes,
		StatusCodesParsed: parsed,
		WildcardForced:    wildcard,
		ExcludeText:       excludeText,
	}
	ret.Headers = optheaders
	return ret
}

func (d *GobusterDir) get(url string) (statusCode *int, size int64, header http.Header, body []byte, err error) {
	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	statusCode, size, header, _, err = d.http.Request(url, libgobuster.RequestOptions{})
	if err != nil {
		return nil, 0, nil, nil, err
	}
	return statusCode, size, header, body, err
}

// NewGobusterDir creates a new initialized Buster
func NewGobusterDir(cont context.Context, opts *OptionsDir) (*GobusterDir, error) {
	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	g := GobusterDir{
		options: opts,
	}

	basicOptions := libgobuster.BasicHTTPOptions{
		Proxy:     opts.Proxy,
		Timeout:   opts.Timeout,
		UserAgent: opts.UserAgent,
	}

	httpOpts := libgobuster.HTTPOptions{
		BasicHTTPOptions: basicOptions,
		FollowRedirect:   opts.FollowRedirect,
		NoTLSValidation:  opts.NoTLSValidation,
		Username:         opts.Username,
		Password:         opts.Password,
		Headers:          opts.Headers,
		Method:           opts.Method,
	}

	h, err := libgobuster.NewHTTPClient(cont, &httpOpts)
	if err != nil {
		return nil, err
	}
	g.http = h
	return &g, nil
}

// PreRun is the pre run implementation of gobusterdir
func (d *GobusterDir) PreRun() error {
	// add trailing slash
	if !strings.HasSuffix(d.options.URL, "/") {
		d.options.URL = fmt.Sprintf("%s/", d.options.URL)
	}

	_, _, _, _, err := d.get(d.options.URL)
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %v", d.options.URL, err)
	}

	guid := uuid.New()
	url := fmt.Sprintf("%s%s", d.options.URL, guid)
	wildcardResp, _, _, _, err := d.get(url)
	if err != nil {
		return err
	}

	if d.options.StatusCodesParsed.Length() > 0 {
		if d.options.StatusCodesParsed.Contains(*wildcardResp) && !d.options.WildcardForced {
			return fmt.Errorf("the server returns a status code that matches the provided options for non existing urls. %s => %d", url, *wildcardResp)
		}
	} else {
		return fmt.Errorf("StatusCodes is not set which should not happen")
	}

	return nil
}

// RunWord is the process implementation of gobusterdir
func (d *GobusterDir) RunWord(word string) (*GoBusterResult, error) {

	// Try the DIR first
	url := fmt.Sprintf("%s%s", d.options.URL, word)
	dirResp, dirSize, _, bodyBytes, err := d.get(url)
	body := string(bodyBytes)
	if err != nil {
		return nil, err
	}
	if dirResp != nil {
		resultStatus := false

		if d.options.StatusCodesParsed.Length() > 0 {
			if d.options.StatusCodesParsed.Contains(*dirResp) {
				resultStatus = true
			}
		} else {
			return nil, fmt.Errorf("StatusCodes is not set which should not happen")
		}
		if resultStatus && (d.options.ExcludeText == "" || !strings.Contains(body, d.options.ExcludeText)) {
			return d.ResultToStruct(word, *dirResp, dirSize), nil
		}
	}
	return nil, nil
}

// Run processes a wordlist and calls RunWord repeatitvely
func (d *GobusterDir) Run(wordlist []string) []GoBusterResult {
	var ret []GoBusterResult
	for _, word := range wordlist {
		res, err := d.RunWord(word)
		if err == nil && res != nil {
			ret = append(ret, *res)
		}
	}
	return ret
}

// ResultToStruct is the to struct implementation of gobusterdir
func (d *GobusterDir) ResultToStruct(entity string, statusCode int, size int64) *GoBusterResult {
	return NewGoBusterResult(entity, statusCode, int(size))
}
