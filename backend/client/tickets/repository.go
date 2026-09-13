package tickets

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDepartmentUnavailable = errors.New("department unavailable")
	ErrLocationUnavailable   = errors.New("location unavailable")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) List(ctx context.Context, query ListQuery) (ListResponse, error) {
	if repository == nil || repository.pool == nil {
		return ListResponse{}, fmt.Errorf("ticket repository is not configured")
	}

	rows, err := repository.pool.Query(ctx, `
SELECT
    t.id,
    t.title,
    t.description,
    t.status::text,
    t.priority,
    t.due_at,
    t.created_at,
    t.updated_at,
    requester.id,
    requester.full_name,
    requester_department.code,
    destination.id,
    destination.code,
    destination.name,
    location.id,
    location.code,
    location.name,
    accepted_user.id,
    accepted_user.full_name,
    accepted_department.code,
    t.accepted_at,
    t.closed_at
FROM tickets AS t
JOIN users AS requester ON requester.id = t.requester_id
JOIN departments AS requester_department ON requester_department.id = requester.department_id
JOIN departments AS destination ON destination.id = t.department_id
LEFT JOIN locations AS location ON location.id = t.location_id
LEFT JOIN users AS accepted_user ON accepted_user.id = t.accepted_by
LEFT JOIN departments AS accepted_department ON accepted_department.id = accepted_user.department_id
WHERE (
    ($1 = 'open' AND t.status IN ('pending'::ticket_status, 'accepted'::ticket_status))
    OR ($1 = 'closed' AND t.status = 'closed'::ticket_status)
)
AND (
    ($2::timestamptz IS NULL AND $3::bigint IS NULL)
    OR (t.created_at, t.id) < ($2::timestamptz, $3::bigint)
)
ORDER BY t.created_at DESC, t.id DESC
LIMIT $4`, string(query.View), query.BeforeCreatedAt, query.BeforeID, query.Limit+1)
	if err != nil {
		return ListResponse{}, fmt.Errorf("query tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]Ticket, 0, query.Limit)
	for rows.Next() {
		var (
			ticket                 Ticket
			status                 string
			acceptedDepartmentCode *string
			locationID             *int64
			locationCode           *string
			locationName           *string
			acceptedID             *int64
			acceptedName           *string
		)
		if err := rows.Scan(
			&ticket.ID,
			&ticket.Title,
			&ticket.Description,
			&status,
			&ticket.Priority,
			&ticket.DueAt,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
			&ticket.Requester.ID,
			&ticket.Requester.FullName,
			&ticket.Requester.DepartmentCode,
			&ticket.Department.ID,
			&ticket.Department.Code,
			&ticket.Department.Name,
			&locationID,
			&locationCode,
			&locationName,
			&acceptedID,
			&acceptedName,
			&acceptedDepartmentCode,
			&ticket.AcceptedAt,
			&ticket.ClosedAt,
		); err != nil {
			return ListResponse{}, fmt.Errorf("scan ticket: %w", err)
		}
		ticket.Status = status
		if locationID != nil {
			ticket.Location = &LocationSummary{ID: *locationID, Code: dereferenceString(locationCode), Name: dereferenceString(locationName)}
		}
		if acceptedID != nil {
			ticket.AcceptedBy = &IdentitySummary{ID: *acceptedID, FullName: dereferenceString(acceptedName), DepartmentCode: dereferenceString(acceptedDepartmentCode)}
		}
		ticket.AssignedDepartments = make([]DepartmentSummary, 0)
		ticket.AssignedUsers = make([]IdentitySummary, 0)
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return ListResponse{}, fmt.Errorf("iterate tickets: %w", err)
	}

	hasMore := len(tickets) > query.Limit
	if hasMore {
		tickets = tickets[:query.Limit]
	}
	response := ListResponse{
		Tickets: tickets,
		Page:    Page{HasMore: hasMore},
	}
	if hasMore && len(tickets) > 0 {
		last := tickets[len(tickets)-1]
		response.Page.NextBeforeCreatedAt = &last.CreatedAt
		response.Page.NextBeforeID = &last.ID
	}
	if len(tickets) == 0 {
		return response, nil
	}

	ticketIDs := make([]int64, len(tickets))
	for index := range tickets {
		ticketIDs[index] = tickets[index].ID
	}
	if err := repository.loadDepartmentAssignments(ctx, tickets, ticketIDs); err != nil {
		return ListResponse{}, err
	}
	if err := repository.loadUserAssignments(ctx, tickets, ticketIDs); err != nil {
		return ListResponse{}, err
	}
	return response, nil
}

func (repository *Repository) Create(ctx context.Context, requester IdentitySummary, input CreateRequest) (Ticket, error) {
	if repository == nil || repository.pool == nil {
		return Ticket{}, fmt.Errorf("ticket repository is not configured")
	}

	transaction, err := repository.pool.Begin(ctx)
	if err != nil {
		return Ticket{}, fmt.Errorf("begin ticket creation: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var department DepartmentSummary
	if err := transaction.QueryRow(ctx, `
SELECT id, code, name
FROM departments
WHERE id = $1 AND is_active = TRUE
FOR SHARE`, input.DepartmentID).Scan(&department.ID, &department.Code, &department.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Ticket{}, ErrDepartmentUnavailable
		}
		return Ticket{}, fmt.Errorf("validate ticket department: %w", err)
	}

	var location *LocationSummary
	if input.LocationID != nil {
		var locationCode *string
		var locationName string
		var locationID int64
		if err := transaction.QueryRow(ctx, `
SELECT id, code, name
FROM locations
WHERE id = $1 AND is_active = TRUE
FOR SHARE`, *input.LocationID).Scan(&locationID, &locationCode, &locationName); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Ticket{}, ErrLocationUnavailable
			}
			return Ticket{}, fmt.Errorf("validate ticket location: %w", err)
		}
		location = &LocationSummary{ID: locationID, Code: dereferenceString(locationCode), Name: locationName}
	}

	var locationValue any
	if input.LocationID != nil {
		locationValue = *input.LocationID
	}
	var descriptionValue any
	if input.Description != nil {
		descriptionValue = *input.Description
	}
	var dueAtValue any
	if input.DueAt != nil {
		dueAtValue = *input.DueAt
	}

	var ticket Ticket
	var status string
	if err := transaction.QueryRow(ctx, `
INSERT INTO tickets (
    requester_id,
    department_id,
    location_id,
    title,
    description,
    status,
    priority,
    due_at,
    accepted_by,
    accepted_at,
    closed_by,
    closed_at
)
VALUES ($1, $2, $3, $4, $5, 'pending'::ticket_status, $6, $7, NULL, NULL, NULL, NULL)
RETURNING id, status::text, priority, due_at, created_at, updated_at`,
		requester.ID,
		department.ID,
		locationValue,
		input.Title,
		descriptionValue,
		input.Priority,
		dueAtValue,
	).Scan(&ticket.ID, &status, &ticket.Priority, &ticket.DueAt, &ticket.CreatedAt, &ticket.UpdatedAt); err != nil {
		return Ticket{}, fmt.Errorf("insert ticket: %w", err)
	}

	ticket.Title = input.Title
	ticket.Description = input.Description
	ticket.Status = status
	ticket.Requester = requester
	ticket.Department = department
	ticket.Location = location
	ticket.AssignedDepartments = make([]DepartmentSummary, 0)
	ticket.AssignedUsers = make([]IdentitySummary, 0)
	if _, err := transaction.Exec(ctx, `
INSERT INTO ticket_activity (ticket_id, actor_user_id, action)
VALUES ($1, $2, 'created')`, ticket.ID, requester.ID); err != nil {
		return Ticket{}, fmt.Errorf("insert ticket creation activity: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return Ticket{}, fmt.Errorf("commit ticket creation: %w", err)
	}
	return ticket, nil
}

func (repository *Repository) loadDepartmentAssignments(ctx context.Context, tickets []Ticket, ticketIDs []int64) error {
	rows, err := repository.pool.Query(ctx, `
SELECT
    assignment.ticket_id,
    department.id,
    department.code,
    department.name
FROM ticket_assigned_departments AS assignment
JOIN departments AS department ON department.id = assignment.department_id
WHERE assignment.ticket_id = ANY($1::bigint[])
ORDER BY assignment.ticket_id ASC, assignment.id ASC`, ticketIDs)
	if err != nil {
		return fmt.Errorf("query ticket department assignments: %w", err)
	}
	defer rows.Close()

	byTicketID := make(map[int64]int, len(tickets))
	for index := range tickets {
		byTicketID[tickets[index].ID] = index
	}
	for rows.Next() {
		var ticketID int64
		var department DepartmentSummary
		if err := rows.Scan(&ticketID, &department.ID, &department.Code, &department.Name); err != nil {
			return fmt.Errorf("scan ticket department assignment: %w", err)
		}
		index, ok := byTicketID[ticketID]
		if ok {
			tickets[index].AssignedDepartments = append(tickets[index].AssignedDepartments, department)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate ticket department assignments: %w", err)
	}
	return nil
}

func (repository *Repository) loadUserAssignments(ctx context.Context, tickets []Ticket, ticketIDs []int64) error {
	rows, err := repository.pool.Query(ctx, `
SELECT
    assignment.ticket_id,
    assigned_user.id,
    assigned_user.full_name,
    assigned_department.code
FROM ticket_assigned_users AS assignment
JOIN users AS assigned_user ON assigned_user.id = assignment.user_id
JOIN departments AS assigned_department ON assigned_department.id = assigned_user.department_id
WHERE assignment.ticket_id = ANY($1::bigint[])
ORDER BY assignment.ticket_id ASC, assignment.id ASC`, ticketIDs)
	if err != nil {
		return fmt.Errorf("query ticket user assignments: %w", err)
	}
	defer rows.Close()

	byTicketID := make(map[int64]int, len(tickets))
	for index := range tickets {
		byTicketID[tickets[index].ID] = index
	}
	for rows.Next() {
		var ticketID int64
		var user IdentitySummary
		if err := rows.Scan(&ticketID, &user.ID, &user.FullName, &user.DepartmentCode); err != nil {
			return fmt.Errorf("scan ticket user assignment: %w", err)
		}
		index, ok := byTicketID[ticketID]
		if ok {
			tickets[index].AssignedUsers = append(tickets[index].AssignedUsers, user)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate ticket user assignments: %w", err)
	}
	return nil
}

func dereferenceString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
