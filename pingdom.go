package main

import (
	"fmt"
	"os"
	"sort"
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

func getUptimes() {
	gotenv.Load()
	client := pingdom.NewClient(os.Getenv("PINGDOM_EMAIL"), os.Getenv("PINGDOM_PASSWORD"), os.Getenv("PINGDOM_TOKEN"))

	checks, err := client.Checks.List()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var res []UptimeResult

	var totalUptime int64
	var totalTime int64

	for i, check := range checks {

		if check.Paused {
			continue
		}

		if check.Status == "paused" {
			continue
		}

		states, err := outageSummary(client, check.ID)
		if err != nil {
			fmt.Println(err)
			continue
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
			continue
		}

		totalUptime += uptimeSec
		totalTime += total

		res = append(res, UptimeResult{
			check:   check,
			uptime:  (float64(uptimeSec) / float64(total)) * 100,
			up:      uptimeSec,
			down:    downtimeSec,
			unknown: unknownSec,
		})
		fmt.Printf("%d of %d\n", i, len(checks))
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].uptime > res[j].uptime
	})

	for _, r := range res {
		downtime := r.down + r.unknown
		dur := time.Second * time.Duration(downtime)

		fmt.Printf("%0.3f | %s (%s) | %s\n", r.uptime, r.check.Name, r.check.Hostname, dur)
	}

	fmt.Printf("\nTotal uptime: %0.3f\n", (float64(totalUptime)/float64(totalTime))*100)

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
