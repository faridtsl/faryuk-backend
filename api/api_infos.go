package api

import (
	"fmt"
	"net/http"
	"time"

	"FaRyuk/api/utils"
	"FaRyuk/internal/types"
	"FaRyuk/models"
)

func getInfos(w http.ResponseWriter, r *http.Request) {
	dbHandler := models.NewDBHandler()
	defer dbHandler.CloseCommentDBConnection()

	results := types.Infos{}

	historyRecords := dbHandler.GetHistoryRecords()

	failedScans := 0
	successfulScans := 0
	ongoingScans := 0
	for _, h := range historyRecords {
		if h.IsFinished && h.IsSuccess {
			successfulScans++
		} else if h.IsFinished {
			failedScans++
		} else {
			ongoingScans++
		}
	}

	results.Failed = fmt.Sprintf("%d", failedScans)
	results.Successful = fmt.Sprintf("%d", successfulScans)
	results.OnGoing = fmt.Sprintf("%d", ongoingScans)

	duration := time.Since(startTime)
	days := int(duration.Hours() / 24)
	hours := int(int(duration.Hours()) - days*24)
	minutes := int(duration.Minutes()) - int(duration.Hours())*60
	uptime := fmt.Sprintf("%dd%0dh%0dm", days, hours, minutes)

	results.Uptime = uptime

	utils.ReturnSuccess(&w, results)
}
