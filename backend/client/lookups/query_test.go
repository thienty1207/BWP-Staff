package lookups

import "testing"

func TestParseLocationQueryValuesLeavesOmittedLimitUnbounded(t *testing.T) {
	query, err := parseLocationQueryValues(map[string]string{})
	if err != nil {
		t.Fatalf("parse default location query: %v", err)
	}
	if query.Search != "" || query.Limit != 0 || query.HasLimit {
		t.Fatalf("unexpected default location query: %+v", query)
	}
}

func TestParseLocationQueryValuesTrimsSearchAndAcceptsExplicitLimit(t *testing.T) {
	query, err := parseLocationQueryValues(map[string]string{"q": "  Server  ", "limit": "30"})
	if err != nil {
		t.Fatalf("parse valid location query: %v", err)
	}
	if query.Search != "Server" || query.Limit != 30 || !query.HasLimit {
		t.Fatalf("unexpected parsed location query: %+v", query)
	}
}

func TestParseLocationQueryValuesRejectsInvalidLimitsAndOversizedSearch(t *testing.T) {
	cases := []map[string]string{
		{"limit": "0"},
		{"limit": "31"},
		{"limit": "many"},
		{"q": "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901"},
	}
	for _, values := range cases {
		if _, err := parseLocationQueryValues(values); err == nil {
			t.Fatalf("expected invalid location query to be rejected: %#v", values)
		}
	}
}
