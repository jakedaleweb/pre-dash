package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
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

	tmpl := template.Must(template.ParseFiles("templates/pingdom.html"))
	data := PingdomPage{
		Title:       "Availability report",
		CwpRes:      parseResults(cwpRes, 99.7),
		SspRes:      parseResults(sspRes, 99.9),
		SspUptime:   fmt.Sprintf("%0.2f", sspUptime),
		CwpUptime:   fmt.Sprintf("%0.2f", cwpUptime),
		SspDiff:     fmt.Sprintf("%0.2f", sspDiff),
		SspIncrease: sspIncrease,
		CwpDiff:     fmt.Sprintf("%0.2f", cwpDiff),
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

	upTimeResults := make(chan UptimeResult)

	for _, check := range checks {
		go func(c pingdom.CheckResponse) {
			calculateUptimeResult(c, client, from, to, upTimeResults)
			wg.Done()
		}(check)
	}

	go func() {
		wg.Wait()
		close(upTimeResults)
	}()

	for result := range upTimeResults {
		switch result.platform {
		case PlatformCWP:
			cwpTotalUptime += result.up
			cwpTotalTime += result.total
			cwpRes = append(cwpRes, result)

		case PlatformSSP:
			sspTotalUptime += result.up
			sspTotalTime += result.total
			sspRes = append(sspRes, result)
		}
	}

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

func calculateUptimeResult(check pingdom.CheckResponse, client *pingdom.Client, from, to time.Time, out chan UptimeResult) {
	if check.Paused {
		return
	}

	if check.Status == "paused" {
		return
	}

	states, err := outageSummary(client, check.ID, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during outageSummary: %s\n", err)
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

	result := UptimeResult{
		check:   check,
		uptime:  (float64(uptimeSec) / float64(total)) * 100,
		up:      uptimeSec,
		down:    downtimeSec,
		unknown: unknownSec,
		total:   total,
	}

	if strings.Contains(check.Name, "CWP") {
		result.platform = PlatformCWP
	} else if strings.Contains(check.Name, "SSP") {
		result.platform = PlatformSSP
	} else {
		return
	}

	out <- result

}

func parseResults(res []UptimeResult, sla float64) []ResultRow {
	var result []ResultRow
	for _, r := range res {
		// calculate how many seconds are allowed to be down
		allowedDowntimeSLA := (100.0 - sla) / 100.0
		allowed := float64(r.total) * allowedDowntimeSLA

		downtime := time.Second * time.Duration(r.down)

		// calculate how much is left in the error budget
		errorBudget := time.Second * time.Duration(int64(allowed)-r.down)

		// only show checks that has less than 15mins in their error budgets
		if errorBudget > 15*time.Minute {
			continue
		}
		row := ResultRow{
			Availability: fmt.Sprintf("%0.2f", r.uptime),
			Name:         r.check.Name,
			Downtime:     fmtDuration(downtime),
			ErrorBudget:  fmtDuration(errorBudget),
			IsMinus:      errorBudget <= 0,
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

// format duration into a string rounded to hours and minutes with a sign
func fmtDuration(dur time.Duration) string {

	var sign string
	if dur < 0 {
		sign = "-"
		dur *= -1
	}

	d := dur.Round(time.Minute)
	h := d / time.Hour

	d -= h * time.Hour
	m := d / time.Minute

	if int(h) != 0 {
		return fmt.Sprintf("%s%dh %02dm", sign, h, m)
	}

	return fmt.Sprintf("%s%2dm", sign, m)
}
