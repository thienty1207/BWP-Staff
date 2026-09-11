package tickets

import (
	"testing"
	"time"
)

func TestParseListQueryValuesUsesSafeDefaultsAndCursor(t *testing.T) {
	query, err := parseListQueryValues(map[string]string{
		"view":              "closed",
		"limit":             "25",
		"before_created_at": "2026-09-11T10:00:00Z",
		"before_id":         "42",
	})
	if err != nil {
		t.Fatalf("parse valid ticket list query: %v", err)
	}

	if query.View != TicketViewClosed {
		t.Fatalf("expected closed view, got %q", query.View)
	}
	if query.Limit != 25 {
		t.Fatalf("expected limit 25, got %d", query.Limit)
	}
	if query.BeforeCreatedAt == nil || !query.BeforeCreatedAt.Equal(time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected before_created_at: %v", query.BeforeCreatedAt)
	}
	if query.BeforeID == nil || *query.BeforeID != 42 {
		t.Fatalf("unexpected before_id: %v", query.BeforeID)
	}
}

func TestParseListQueryValuesDefaultsToOpenAndFifty(t *testing.T) {
	query, err := parseListQueryValues(map[string]string{})
	if err != nil {
		t.Fatalf("parse default ticket list query: %v", err)
	}
	if query.View != TicketViewOpen || query.Limit != 50 {
		t.Fatalf("unexpected defaults: view=%q limit=%d", query.View, query.Limit)
	}
	if query.BeforeCreatedAt != nil || query.BeforeID != nil {
		t.Fatalf("default query unexpectedly has a cursor: %+v", query)
	}
}

func TestParseListQueryValuesRejectsInvalidParameters(t *testing.T) {
	cases := []struct {
		name   string
		values map[string]string
	}{
		{name: "invalid view", values: map[string]string{"view": "all"}},
		{name: "zero limit", values: map[string]string{"limit": "0"}},
		{name: "limit too large", values: map[string]string{"limit": "101"}},
		{name: "non numeric limit", values: map[string]string{"limit": "many"}},
		{name: "partial cursor timestamp", values: map[string]string{"before_created_at": "2026-09-11T10:00:00Z"}},
		{name: "partial cursor id", values: map[string]string{"before_id": "42"}},
		{name: "invalid timestamp", values: map[string]string{"before_created_at": "not-a-time", "before_id": "42"}},
		{name: "invalid cursor id", values: map[string]string{"before_created_at": "2026-09-11T10:00:00Z", "before_id": "0"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := parseListQueryValues(testCase.values); err == nil {
				t.Fatal("expected invalid ticket list query to be rejected")
			}
		})
	}
}
