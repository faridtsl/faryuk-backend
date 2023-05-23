package types

import (
	"time"

	"FaRyuk/pkg"
)

// WebResult : struct for webresults of a port
type WebResult struct {
	Port         int                  `bson:"port" json:"port"`
	Ssl          bool                 `bson:"ssl" json:"ssl"`
	Headers      map[string][]string  `bson:"headers" json:"headers"`
	Busterres    []pkg.GoBusterResult `bson:"busterres" json:"busterres"`
	Screen       pkg.ScreenerResult   `bson:"screener" json:"screen"`
	RunnerOutput []RunnerResult       `bson:"runnerOutput" json:"runnerOutput"`
	CreatedDate  time.Time            `bson:"createdDate" json:"createdDate"`
	Err          []string             `bson:"err" json:"err"`
}

// Result : struct for result of a host
type Result struct {
	ID           string         `bson:"id" json:"id"`
	Host         string         `bson:"host" json:"host"`
	Ips          []string       `bson:"ips" json:"ips"`
	OpenPorts    []int          `bson:"openPorts" json:"openPorts"`
	WebResults   []WebResult    `bson:"webResults" json:"webResults"`
	Tags         []string       `bson:"tags" json:"tags"`
	RunnerOutput []RunnerResult `bson:"runnerOutput" json:"runnerOutput"`
	Owner        string         `bson:"owner" json:"owner"`
	SharedWith   []string       `bson:"sharedWith" json:"sharedWith"`
	OwnerGroup   string         `bson:"ownerGroup" json:"ownerGroup"`
	CreatedDate  time.Time      `bson:"createdDate" json:"createdDate"`
	Err          []string       `bson:"err" json:"err"`
}

// Runner
type Runner struct {
	ID          string   `bson:"id" json:"id"`
	Tag         string   `bson:"tag" json:"tag"`
	DisplayName string   `bson:"displayName" json:"displayName"`
	Cmd         []string `bson:"cmd" json:"cmd"`
	IsWeb       bool     `bson:"isWeb" json:"isWeb"`
	IsPort      bool     `bson:"isPort" json:"isPort"`
	Owner       string   `bson:"owner" json:"owner"`
}

// RunnerResult : result from docker tool
type RunnerResult struct {
	ID          string `bson:"id" json:"id"`
	ToolName    string `bson:"toolName" json:"toolName"`
	ScannedPort string `bson:"scannedPort" json:"scannedPort"`
	Output      string `bson:"output" json:"output"`
	Stderr      string `bson:"stderr" json:"stderr"`
}

// HistoryRecord : record of the history of a scan
type HistoryRecord struct {
	ID          string    `bson:"id" json:"id"`
	Host        string    `bson:"host" json:"host"`
	Domain      string    `bson:"domain" json:"domain"`
	IsWeb       bool      `bson:"isWeb" json:"isWeb"`
	IsFinished  bool      `bson:"isFinished" json:"isFinished"`
	IsSuccess   bool      `bson:"isSuccess" json:"isSuccess"`
	State       []string  `bson:"state" json:"state"`
	Owner       string    `bson:"owner" json:"owner"`
	OwnerGroup  string    `bson:"ownerGroup" json:"ownerGroup"`
	CreatedDate time.Time `bson:"createdDate" json:"createdDate"`
}

// Comment : struct for comment on a result
type Comment struct {
	ID               string    `bson:"id" json:"id"`
	Content          string    `bson:"content" json:"content"`
	ImageAttachement string    `bson:"imageAttachement" json:"imageAttachement"`
	IDResult         string    `bson:"idResult" json:"idResult"`
	Owner            string    `bson:"owner" json:"owner"`
	CreatedDate      time.Time `bson:"createdDate" json:"createdDate"`
	UpdatedDate      time.Time `bson:"updatedDate" json:"updatedDate"`
}

// User : user struct
type User struct {
	ID       string  `bson:"id" json:"id"`
	Username string  `bson:"username" json:"username"`
	Password string  `bson:"password" json:"password"`
	Theme    string  `bson:"theme" json:"theme"`
	Groups   []Group `bson:"groups" json:"groups"`
}

// Group : workgroup struct to permit result sharing
type Group struct {
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

// Sharing : struct to share results between users
type Sharing struct {
	ID       string `bson:"id" json:"id"`
	UserID   string `bson:"userId" json:"userId"`
	OwnerID  string `bson:"ownerId" json:"ownerId"`
	ResultID string `bson:"resultId" json:"resultId"`
	State    string `bson:"state" json:"state"`
}

// App infos : struct for app infos (scans and uptime)
type Infos struct {
	Uptime     string `bson:"uptime" json:"uptime"`
	OnGoing    string `bson:"onGoing" json:"onGoing"`
	Successful string `bson:"successful" json:"successful"`
	Failed     string `bson:"failed" json:"failed"`
}

// JSONReturn : struct used by the API to normalize everything
type JSONReturn struct {
	Status string      `bson:"status" json:"status"`
	Body   interface{} `bson:"body" json:"body"`
}
