package main

import (
	"time"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

type Tickets []Ticket

type Ticket struct {
	CcEmails        []interface{} `json:"cc_emails"`
	FwdEmails       []interface{} `json:"fwd_emails"`
	ReplyCcEmails   []interface{} `json:"reply_cc_emails"`
	FrEscalated     bool          `json:"fr_escalated"`
	Spam            bool          `json:"spam"`
	EmailConfigID   int64         `json:"email_config_id"`
	GroupID         int64         `json:"group_id"`
	Priority        int           `json:"priority"`
	RequesterID     int64         `json:"requester_id"`
	ResponderID     interface{}   `json:"responder_id"`
	Source          int           `json:"source"`
	CompanyID       int64         `json:"company_id"`
	Status          int           `json:"status"`
	Subject         string        `json:"subject"`
	AssociationType interface{}   `json:"association_type"`
	ToEmails        []string      `json:"to_emails"`
	ProductID       int64         `json:"product_id"`
	ID              int           `json:"id"`
	Type            string        `json:"type"`
	DueBy           time.Time     `json:"due_by"`
	FrDueBy         time.Time     `json:"fr_due_by"`
	IsEscalated     bool          `json:"is_escalated"`
	Description     string        `json:"description"`
	DescriptionText string        `json:"description_text"`
	CustomFields    struct {
		CfClosureCode             interface{} `json:"cf_closure_code"`
		Estimate                  interface{} `json:"estimate"`
		CfCodeCare                bool        `json:"cf_code_care"`
		CfNextScheduledAction     interface{} `json:"cf_next_scheduled_action"`
		CfFlagForTeamDiscussion   bool        `json:"cf_flag_for_team_discussion"`
		CfEstimatedCompletionDate interface{} `json:"cf_estimated_completion_date"`
	} `json:"custom_fields"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
	AssociatedTicketsCount interface{}   `json:"associated_tickets_count"`
	Tags                   []interface{} `json:"tags"`
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

type UptimeResult struct {
	check   pingdom.CheckResponse
	uptime  float64
	up      int64
	down    int64
	unknown int64
}

type HomePage struct {
	Title string
	Body  []byte
}

type PingdomPage struct {
	Title     string
	CwpRes    []ResultRow
	SspRes    []ResultRow
	SspUptime string
	CwpUptime string
}

type FreshdeskPage struct {
	Title string
	Avg   string
	Count int
}

type ResultRow struct {
	Availability string
	Name         string
	Downtime     string
}
