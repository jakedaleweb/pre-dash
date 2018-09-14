package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/russellcardullo/go-pingdom/pingdom"
	"github.com/subosito/gotenv"
)

type UptimeResult struct {
	check   pingdom.CheckResponse
	uptime  float64
	up      int64
	down    int64
	unknown int64
}

func getUptimes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=300")
	gotenv.Load()
	client := pingdom.NewClient(os.Getenv("PINGDOM_EMAIL"), os.Getenv("PINGDOM_PASSWORD"), os.Getenv("PINGDOM_TOKEN"))

	checks, err := client.Checks.List()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var cwpRes []UptimeResult
	var sspRes []UptimeResult

	var cwpTotalUptime int64
	var cwpTotalTime int64

	var sspTotalUptime int64
	var sspTotalTime int64

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

			states, err := outageSummary(client, check.ID)
			if err != nil {
				fmt.Println(err)
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

	sort.Slice(cwpRes, func(i, j int) bool {
		return cwpRes[i].uptime > cwpRes[j].uptime
	})

	sort.Slice(sspRes, func(i, j int) bool {
		return sspRes[i].uptime > sspRes[j].uptime
	})

	for _, r := range cwpRes {
		downtime := r.down + r.unknown
		dur := time.Second * time.Duration(downtime)

		uptimeString := strconv.FormatFloat(r.uptime, 'f', 3, 64)
		cwpUptimeMessage := strings.Join([]string{uptimeString, r.check.Name, dur.String(), "\n"}, " | ")
		w.Write([]byte(cwpUptimeMessage))
	}

	w.Write([]byte("\n"))

	for _, r := range sspRes {
		downtime := r.down + r.unknown
		dur := time.Second * time.Duration(downtime)

		uptimeString := strconv.FormatFloat(r.uptime, 'f', 3, 64)
		sspUptimeMessage := strings.Join([]string{uptimeString, r.check.Name, dur.String(), "\n"}, " | ")
		w.Write([]byte(sspUptimeMessage))
	}

	cwpUptime := strconv.FormatFloat((float64(cwpTotalUptime)/float64(cwpTotalTime))*100, 'f', 3, 64)
	cwpMessage := strings.Join([]string{"CWP Total uptime:", cwpUptime, "\n"}, " ")

	sspUptime := strconv.FormatFloat((float64(sspTotalUptime)/float64(sspTotalTime))*100, 'f', 3, 64)
	sspMessage := strings.Join([]string{"SSP Total uptime:", sspUptime, "\n"}, " ")

	w.Write([]byte("\n"))

	w.Write([]byte(cwpMessage))
	w.Write([]byte(sspMessage))
}

type summaryOutageJsonResponse struct {
	Summary struct {
		States []State `json:"states"`
	} `json:"summary"`
}

type State struct {
	Status   string `json:"status"`
	Timefrom int64  `json:"timefrom"`
	Timeto   int64  `json:"timeto"`
}

func (s *State) From() time.Time {
	return time.Unix(s.Timefrom, 0)
}

func (s *State) To() time.Time {
	return time.Unix(s.Timeto, 0)
}

func outageSummary(client *pingdom.Client, checkid int) ([]State, error) {

	from := time.Now().Add(-14 * 24 * time.Hour)
	params := make(map[string]string)
	params["from"] = fmt.Sprintf("%d", from.Unix())

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
