package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

func getUptimes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	from := time.Now().Add(-14 * 24 * time.Hour)
	to := time.Now()
	cwpRes, sspRes, sspUptime, cwpUptime, err := formatSummaries(from, to)
	if err != nil {
		log.Fatal(err)
	}

	fromPrev := time.Now().Add(-28 * 24 * time.Hour)
	toPrev := time.Now().Add(-14 * 24 * time.Hour)
	_, _, sspUptimePrev, cwpUptimePrev, prevErr := formatSummaries(fromPrev, toPrev)
	if prevErr != nil {
		log.Fatal(err)
	}

	cwpUptimeStr := strconv.FormatFloat(cwpUptime, 'f', 3, 64)
	sspUptimeStr := strconv.FormatFloat(sspUptime, 'f', 3, 64)

	cwpDiff := cwpUptime - cwpUptimePrev
	// Check if the diff is less than 0 (i.e. not increased)
	cwpIncrease := true
	if cwpDiff < 0 {
		cwpIncrease = false
	}

	sspDiff := sspUptime - sspUptimePrev
	// Check if the diff is less than 0 (i.e. not increased)
	sspIncrease := true
	if sspDiff < 0 {
		sspIncrease = false
	}

	cwpDiffStr := strconv.FormatFloat(cwpDiff, 'f', 3, 64)
	sspDiffStr := strconv.FormatFloat(sspDiff, 'f', 3, 64)

	tmpl := template.Must(template.ParseFiles("templates/pingdom.html"))
	data := PingdomPage{
		Title:       "Availability report",
		CwpRes:      parseResults(cwpRes, 99.7),
		SspRes:      parseResults(sspRes, 99.9),
		SspUptime:   sspUptimeStr,
		CwpUptime:   cwpUptimeStr,
		SspDiff:     sspDiffStr,
		SspIncrease: sspIncrease,
		CwpDiff:     cwpDiffStr,
		CwpIncrease: cwpIncrease,
	}
	tmpl.Execute(w, data)
}

func formatSummaries(from time.Time, to time.Time) ([]UptimeResult, []UptimeResult, float64, float64, error) {

	var cwpRes []UptimeResult
	var sspRes []UptimeResult

	var cwpTotalUptime int64
	var cwpTotalTime int64

	var sspTotalUptime int64
	var sspTotalTime int64

	client := pingdom.NewClient(os.Getenv("PINGDOM_EMAIL"), os.Getenv("PINGDOM_PASSWORD"), os.Getenv("PINGDOM_TOKEN"))

	checks, err := client.Checks.List()
	if err != nil {
		return nil, nil, 0, 0, err
	}

	var wg sync.WaitGroup
	wg.Add(len(checks))

	for i, check := range checks {
		go func(i int, check pingdom.CheckResponse) {
			defer wg.Done()

			if check.Paused {
				return
			}

			if check.Status == "paused" {
				return
			}

			states, err := outageSummary(client, check.ID, from, to)
			if err != nil {
				return
			}

			var uptimeSec int64
			var downtimeSec int64
			var unknownSec int64

			for _, k := range states {
				switch k.Status {
				case "up":
					uptimeSec += k.Timeto - k.Timefrom
				case "down":
					downtimeSec += k.Timeto - k.Timefrom
				case "unknown":
					unknownSec += k.Timeto - k.Timefrom
				}
			}

			total := uptimeSec + downtimeSec + unknownSec
			if total == 0 {
				return
			}

			if strings.Contains(check.Name, "CWP") {
				cwpTotalUptime += uptimeSec
				cwpTotalTime += total
				cwpRes = append(cwpRes, UptimeResult{
					check:   check,
					uptime:  (float64(uptimeSec) / float64(total)) * 100,
					up:      uptimeSec,
					down:    downtimeSec,
					unknown: unknownSec,
				})
			}

			if strings.Contains(check.Name, "SSP") {
				sspTotalUptime += uptimeSec
				sspTotalTime += total
				sspRes = append(sspRes, UptimeResult{
					check:   check,
					uptime:  (float64(uptimeSec) / float64(total)) * 100,
					up:      uptimeSec,
					down:    downtimeSec,
					unknown: unknownSec,
				})
			}
		}(i, check)
	}

	wg.Wait()

	cwpUptime := float64(cwpTotalUptime) / float64(cwpTotalTime) * 100
	sspUptime := float64(sspTotalUptime) / float64(sspTotalTime) * 100

	sort.Slice(cwpRes, func(i, j int) bool {
		return cwpRes[i].uptime < cwpRes[j].uptime
	})

	sort.Slice(sspRes, func(i, j int) bool {
		return sspRes[i].uptime < sspRes[j].uptime
	})

	return cwpRes, sspRes, cwpUptime, sspUptime, nil
}

func parseResults(res []UptimeResult, sla float64) []ResultRow {
	var result []ResultRow
	for _, r := range res {

		// Don't display checks that are withing the SLA
		if r.uptime > sla {
			continue
		}

		downtime := r.down + r.unknown
		dur := time.Second * time.Duration(downtime)

		uptimeString := strconv.FormatFloat(r.uptime, 'f', 3, 64)

		row := ResultRow{
			Availability: uptimeString,
			Name:         r.check.Name,
			Downtime:     dur.String(),
		}

		result = append(result, row)
	}

	return result
}

func (s *State) From() time.Time {
	return time.Unix(s.Timefrom, 0)
}

func (s *State) To() time.Time {
	return time.Unix(s.Timeto, 0)
}

func outageSummary(client *pingdom.Client, checkid int, from time.Time, to time.Time) ([]State, error) {

	params := make(map[string]string)
	params["from"] = fmt.Sprintf("%d", from.Unix())
	params["to"] = fmt.Sprintf("%d", to.Unix())

	url := fmt.Sprintf("/summary.outage/%d", checkid)

	req, err := client.NewRequest("GET", url, params)
	if err != nil {
		return nil, err
	}

	var res summaryOutageJsonResponse

	resp, err := client.Do(req, &res)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return res.Summary.States, nil
}
