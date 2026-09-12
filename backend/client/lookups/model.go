package lookups

type Department struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Location struct {
	ID   int64   `json:"id"`
	Code *string `json:"code"`
	Name string  `json:"name"`
}

type DepartmentsResponse struct {
	Departments []Department `json:"departments"`
}

type LocationsResponse struct {
	Locations []Location `json:"locations"`
}
