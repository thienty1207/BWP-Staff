package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thienty1207/BWP-Staff/backend/shared/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("authentication required")
	errUserNotActive      = errors.New("user is no longer active")
)

const defaultSessionTTL = 12 * time.Hour

const dummyPasswordHash = "$argon2id$v=19$m=65536,t=3,p=4$p/yr3By2abs0dKe91VXMmw$avvdWZEbtHf9TRf444e+QwmSMtJAEiHsSm5+1F/UsIg"

type Service struct {
	repository *Repository
	sessionTTL time.Duration
}

func NewService(repository *Repository, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTL
	}
	return &Service{repository: repository, sessionTTL: sessionTTL}
}

func (service *Service) Login(ctx context.Context, input loginRequest, metadata sessionMetadata) (loginResult, error) {
	user, err := service.repository.FindUserByUsername(ctx, input.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = security.VerifyPassword(input.Password, dummyPasswordHash)
		return loginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return loginResult{}, err
	}
	if !verifyLoginPassword(&user, input.Password) || !user.IsActive {
		return loginResult{}, ErrInvalidCredentials
	}

	rawToken, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return loginResult{}, err
	}
	createdAt := time.Now().UTC()
	expiresAt := createdAt.Add(service.sessionTTL)
	if _, err := service.repository.CreateSession(ctx, sessionRecord{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
		UserAgent: metadata.UserAgent,
		IPAddress: metadata.IPAddress,
	}); err != nil {
		if errors.Is(err, errUserNotActive) {
			return loginResult{}, ErrInvalidCredentials
		}
		return loginResult{}, err
	}

	return loginResult{
		User:      identityFromUser(user),
		RawToken:  rawToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (service *Service) Authenticate(ctx context.Context, rawToken string) (Principal, error) {
	if !security.IsValidSessionToken(rawToken) {
		return Principal{}, ErrUnauthenticated
	}

	principal, err := service.repository.FindValidSession(ctx, security.HashSessionToken(rawToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return Principal{}, ErrUnauthenticated
	}
	if err != nil {
		return Principal{}, err
	}
	return principal, nil
}

func (service *Service) Logout(ctx context.Context, rawToken string) error {
	if !security.IsValidSessionToken(rawToken) {
		return nil
	}
	return service.repository.RevokeSession(ctx, security.HashSessionToken(rawToken))
}

func verifyLoginPassword(user *userRecord, password string) bool {
	if user == nil {
		return security.VerifyPassword(password, dummyPasswordHash)
	}
	return security.VerifyPassword(password, user.PasswordHash)
}
