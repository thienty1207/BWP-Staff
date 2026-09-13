package lookups

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	maxLocationLimit     = 30
	maxLocationQuerySize = 100
)

var errInvalidLocationQuery = errors.New("invalid location lookup query")

type LocationQuery struct {
	Search   string
	Limit    int
	HasLimit bool
}

func parseLocationQueryValues(values map[string]string) (LocationQuery, error) {
	search := strings.TrimSpace(values["q"])
	if !utf8.ValidString(search) || utf8.RuneCountInString(search) > maxLocationQuerySize {
		return LocationQuery{}, errInvalidLocationQuery
	}

	query := LocationQuery{Search: search}
	rawLimit, hasLimit := values["limit"]
	if !hasLimit {
		return query, nil
	}
	limit, err := strconv.Atoi(rawLimit)
	if err != nil || limit < 1 || limit > maxLocationLimit {
		return LocationQuery{}, errInvalidLocationQuery
	}
	query.Limit = limit
	query.HasLimit = true
	return query, nil
}
