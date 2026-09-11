package tickets

import (
	"strconv"
	"strings"
	"time"
)

type TicketView string

const (
	TicketViewOpen   TicketView = "open"
	TicketViewClosed TicketView = "closed"

	defaultTicketLimit = 50
	maxTicketLimit     = 100
)

type ListQuery struct {
	View            TicketView
	Limit           int
	BeforeCreatedAt *time.Time
	BeforeID        *int64
}

type ListResponse struct {
	Tickets []Ticket `json:"tickets"`
	Page    Page     `json:"page"`
}

type Page struct {
	HasMore             bool       `json:"has_more"`
	NextBeforeCreatedAt *time.Time `json:"next_before_created_at"`
	NextBeforeID        *int64     `json:"next_before_id"`
}

type Ticket struct {
	ID                  int64               `json:"id"`
	Title               string              `json:"title"`
	Status              string              `json:"status"`
	Priority            bool                `json:"priority"`
	DueAt               *time.Time          `json:"due_at"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	Requester           IdentitySummary     `json:"requester"`
	Department          DepartmentSummary   `json:"department"`
	Location            *LocationSummary    `json:"location"`
	AcceptedBy          *IdentitySummary    `json:"accepted_by"`
	AcceptedAt          *time.Time          `json:"accepted_at"`
	AssignedDepartments []DepartmentSummary `json:"assigned_departments"`
	AssignedUsers       []IdentitySummary   `json:"assigned_users"`
	ClosedAt            *time.Time          `json:"closed_at"`
}

type IdentitySummary struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
}

type DepartmentSummary struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type LocationSummary struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func parseListQueryValues(values map[string]string) (ListQuery, error) {
	query := ListQuery{
		View:  TicketViewOpen,
		Limit: defaultTicketLimit,
	}

	if view, ok := values["view"]; ok {
		switch TicketView(view) {
		case TicketViewOpen, TicketViewClosed:
			query.View = TicketView(view)
		default:
			return ListQuery{}, invalidQueryError("view")
		}
	}

	if limitValue, ok := values["limit"]; ok {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit < 1 || limit > maxTicketLimit {
			return ListQuery{}, invalidQueryError("limit")
		}
		query.Limit = limit
	}

	beforeCreatedAtValue, hasBeforeCreatedAt := values["before_created_at"]
	beforeIDValue, hasBeforeID := values["before_id"]
	if hasBeforeCreatedAt != hasBeforeID {
		return ListQuery{}, invalidQueryError("cursor")
	}
	if hasBeforeCreatedAt {
		beforeCreatedAt, err := time.Parse(time.RFC3339, beforeCreatedAtValue)
		if err != nil {
			return ListQuery{}, invalidQueryError("before_created_at")
		}
		beforeID, err := strconv.ParseInt(beforeIDValue, 10, 64)
		if err != nil || beforeID <= 0 {
			return ListQuery{}, invalidQueryError("before_id")
		}
		beforeCreatedAt = beforeCreatedAt.UTC()
		query.BeforeCreatedAt = &beforeCreatedAt
		query.BeforeID = &beforeID
	}

	return query, nil
}

type invalidQueryError string

func (err invalidQueryError) Error() string {
	return "invalid ticket list query: " + strings.TrimSpace(string(err))
}
