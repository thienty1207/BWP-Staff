package auth

import "time"

const (
	sessionCookieName     = "bwp_session"
	maxUsernameCharacters = 50
	maxPasswordBytes      = 1024
	maxUserAgentBytes     = 512
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userRecord struct {
	ID           int64
	Username     string
	EmployeeCode string
	PasswordHash string
	FullName     string
	Role         string
	AvatarURL    *string
	IsActive     bool
	Department   departmentRecord
}

type departmentRecord struct {
	ID   int64
	Code string
	Name string
}

// UserIdentity is the safe public identity returned by authentication
// endpoints and retained in the request-local principal.
type UserIdentity struct {
	ID           int64              `json:"id"`
	Username     string             `json:"username"`
	EmployeeCode string             `json:"employee_code"`
	FullName     string             `json:"full_name"`
	Role         string             `json:"role"`
	Department   DepartmentIdentity `json:"department"`
	AvatarURL    *string            `json:"avatar_url"`
}

type DepartmentIdentity struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// Principal is the request-scoped authenticated identity. It deliberately
// contains no password or token material.
type Principal struct {
	User             UserIdentity
	SessionID        int64
	SessionExpiresAt time.Time
}

type sessionMetadata struct {
	UserAgent *string
	IPAddress *string
}

type sessionRecord struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	UserAgent *string
	IPAddress *string
}

type loginResult struct {
	User      UserIdentity
	RawToken  string
	ExpiresAt time.Time
}

func identityFromUser(user userRecord) UserIdentity {
	return UserIdentity{
		ID:           user.ID,
		Username:     user.Username,
		EmployeeCode: user.EmployeeCode,
		FullName:     user.FullName,
		Role:         user.Role,
		Department: DepartmentIdentity{
			ID:   user.Department.ID,
			Code: user.Department.Code,
			Name: user.Department.Name,
		},
		AvatarURL: user.AvatarURL,
	}
}
