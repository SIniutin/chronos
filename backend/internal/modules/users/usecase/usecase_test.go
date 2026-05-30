package users_usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

var _ users_api.RegisterUsecase = (*RegisterUsecase)(nil)
var _ users_api.LoginUsecase = (*LoginUsecase)(nil)
var _ users_api.RefreshUsecase = (*RefreshUsecase)(nil)
var _ users_api.LogoutUsecase = (*LogoutUsecase)(nil)
var _ users_api.GetMeUsecase = (*GetMeUsecase)(nil)
var _ users_api.ChangeUserRoleUsecase = (*ChangeUserRoleUsecase)(nil)
var _ users_api.FindUserUsecase = (*FindUserUsecase)(nil)

func TestRegisterUsecase(t *testing.T) {
	repo := newMemoryRepo()
	uc := NewRegisterUsecase(repo, fakeHasher{})

	user, err := uc.Exec(context.Background(), users_api.RegisterInput{
		Email:    "USER@Example.COM",
		Login:    "tester",
		Password: "password1",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.Email != "user@example.com" || user.Login != "tester" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if _, err := uc.Exec(context.Background(), users_api.RegisterInput{
		Email:    "user@example.com",
		Login:    "other",
		Password: "password1",
	}); !errors.Is(err, users_api.ErrConflictEmail) {
		t.Fatalf("expected email conflict, got %v", err)
	}
	if _, err := uc.Exec(context.Background(), users_api.RegisterInput{
		Email:    "other@example.com",
		Login:    "tester",
		Password: "password1",
	}); !errors.Is(err, users_api.ErrConflictLogin) {
		t.Fatalf("expected login conflict, got %v", err)
	}
	if _, err := uc.Exec(context.Background(), users_api.RegisterInput{
		Email:    "bad",
		Login:    "te",
		Password: "short",
	}); !errors.Is(err, users_api.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestLoginUsecaseByEmailAndLogin(t *testing.T) {
	repo := newMemoryRepo()
	register := NewRegisterUsecase(repo, fakeHasher{})
	user, err := register.Exec(context.Background(), users_api.RegisterInput{
		Email:    "user@example.com",
		Login:    "tester",
		Password: "password1",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	tokens := newFakeTokenService()
	login := NewLoginUsecase(repo, repo, fakeHasher{}, tokens)

	byEmail, err := login.Exec(context.Background(), users_api.LoginInput{
		Identity: "user@example.com",
		Password: "password1",
	})
	if err != nil {
		t.Fatalf("login by email failed: %v", err)
	}
	if byEmail.AccessToken == "" || byEmail.RefreshToken == "" {
		t.Fatalf("expected tokens, got %+v", byEmail)
	}

	byLogin, err := login.Exec(context.Background(), users_api.LoginInput{
		Identity: "tester",
		Password: "password1",
	})
	if err != nil {
		t.Fatalf("login by login failed: %v", err)
	}
	if byLogin.AccessToken == byEmail.AccessToken {
		t.Fatalf("expected unique access token")
	}

	if _, err := login.Exec(context.Background(), users_api.LoginInput{
		Identity: "tester",
		Password: "wrongpass",
	}); !errors.Is(err, users_api.ErrInvalidCredential) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if _, err := repo.GetRefresh(context.Background(), tokens.HashRefresh(byEmail.RefreshToken)); err != nil {
		t.Fatalf("expected refresh session for user %v: %v", user.ID, err)
	}
}

func TestFindUserUsecaseByEmailAndLogin(t *testing.T) {
	repo := newMemoryRepo()
	register := NewRegisterUsecase(repo, fakeHasher{})
	_, err := register.Exec(context.Background(), users_api.RegisterInput{
		Email:    "user@example.com",
		Login:    "tester",
		Password: "password1",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	uc := NewFindUserUsecase(repo)
	byEmail, err := uc.Exec(context.Background(), users_api.FindUserInput{Identity: "user@example.com"})
	if err != nil {
		t.Fatalf("find by email failed: %v", err)
	}
	byLogin, err := uc.Exec(context.Background(), users_api.FindUserInput{Identity: "tester"})
	if err != nil {
		t.Fatalf("find by login failed: %v", err)
	}
	if byEmail.ID != byLogin.ID || byEmail.Login != "tester" {
		t.Fatalf("unexpected users: email=%+v login=%+v", byEmail, byLogin)
	}
	if _, err := uc.Exec(context.Background(), users_api.FindUserInput{Identity: "??"}); !errors.Is(err, users_api.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if _, err := uc.Exec(context.Background(), users_api.FindUserInput{Identity: "missing@example.com"}); !errors.Is(err, users_api.ErrUserNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestRefreshUsecaseRotatesSession(t *testing.T) {
	repo := newMemoryRepo()
	tokens := newFakeTokenService()
	userID := ud.UserID(uuid.New())
	repo.users[userID] = ud.User{ID: userID, Role: ud.RoleStudent}
	pair, session, err := tokens.NewPair(userID, ud.RoleStudent)
	if err != nil {
		t.Fatalf("new pair failed: %v", err)
	}
	if err := repo.CreateRefresh(context.Background(), session); err != nil {
		t.Fatalf("create refresh failed: %v", err)
	}

	refresh := NewRefreshUsecase(repo, repo, tokens)
	nextPair, err := refresh.Exec(context.Background(), users_api.RefreshInput{RefreshToken: pair.RefreshToken})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if nextPair.RefreshToken == pair.RefreshToken {
		t.Fatalf("expected rotated refresh token")
	}
	oldSession, err := repo.GetRefresh(context.Background(), tokens.HashRefresh(pair.RefreshToken))
	if err != nil {
		t.Fatalf("expected old session: %v", err)
	}
	if oldSession.RevokedAt == nil {
		t.Fatalf("expected old session to be revoked")
	}
	if _, err := refresh.Exec(context.Background(), users_api.RefreshInput{RefreshToken: pair.RefreshToken}); !errors.Is(err, users_api.ErrSessionRevoked) {
		t.Fatalf("expected revoked session error, got %v", err)
	}
}

func TestLogoutUsecaseRevokesProvidedSession(t *testing.T) {
	repo := newMemoryRepo()
	tokens := newFakeTokenService()
	userID := ud.UserID(uuid.New())
	pair, session, err := tokens.NewPair(userID, ud.RoleStudent)
	if err != nil {
		t.Fatalf("new pair failed: %v", err)
	}
	if err := repo.CreateRefresh(context.Background(), session); err != nil {
		t.Fatalf("create refresh failed: %v", err)
	}

	logout := NewLogoutUsecase(repo, tokens)
	if err := logout.Exec(context.Background(), users_api.LogoutInput{
		UserID:       uuid.UUID(userID).String(),
		RefreshToken: pair.RefreshToken,
	}); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	stored, err := repo.GetRefresh(context.Background(), tokens.HashRefresh(pair.RefreshToken))
	if err != nil {
		t.Fatalf("expected session: %v", err)
	}
	if stored.RevokedAt == nil {
		t.Fatalf("expected revoked session")
	}
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (ud.PasswordHash, error) {
	return ud.PasswordHash{Algo: ud.PasswordHashBcrypt, Value: "hash:" + password}, nil
}

func (fakeHasher) Compare(password string, hash ud.PasswordHash) bool {
	return hash.Algo == ud.PasswordHashBcrypt && hash.Value == "hash:"+password
}

type fakeTokenService struct {
	next int
	now  time.Time
}

func newFakeTokenService() *fakeTokenService {
	return &fakeTokenService{now: time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)}
}

func (s *fakeTokenService) NewPair(userID ud.UserID, role ud.Role) (ud.TokenPair, ud.RefreshSession, error) {
	s.next++
	refresh := fmt.Sprintf("refresh-%d", s.next)
	sessionID := uuid.New()
	expiresAt := s.now.Add(time.Hour)
	return ud.TokenPair{
			AccessToken:           fmt.Sprintf("access-%d-%s", s.next, role),
			RefreshToken:          refresh,
			AccessExpires:         s.now.Add(time.Minute),
			RefreshTokenExpiresAt: expiresAt,
		}, ud.RefreshSession{
			ID:        ud.SessionID(sessionID),
			UserID:    userID,
			TokenHash: s.HashRefresh(refresh),
			CreatedAt: s.now,
			ExpiresAt: expiresAt,
		}, nil
}

func (s *fakeTokenService) HashRefresh(token string) string {
	return "hash:" + token
}

func (s *fakeTokenService) RefreshExpired(expiresAt time.Time) bool {
	return !expiresAt.After(s.now)
}

type memoryRepo struct {
	users    map[ud.UserID]ud.User
	creds    map[ud.UserID]ud.Credentials
	byEmail  map[ud.Email]ud.UserID
	byLogin  map[ud.Login]ud.UserID
	sessions map[string]ud.RefreshSession
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		users:    make(map[ud.UserID]ud.User),
		creds:    make(map[ud.UserID]ud.Credentials),
		byEmail:  make(map[ud.Email]ud.UserID),
		byLogin:  make(map[ud.Login]ud.UserID),
		sessions: make(map[string]ud.RefreshSession),
	}
}

func (r *memoryRepo) Create(_ context.Context, params ud.CreateUserParams) (ud.User, error) {
	if _, ok := r.byEmail[params.Email]; ok {
		return ud.User{}, ud.ErrConflictEmail
	}
	if _, ok := r.byLogin[params.Login]; ok {
		return ud.User{}, ud.ErrConflictLogin
	}
	user := ud.User{
		ID:        params.ID,
		Email:     params.Email,
		Login:     params.Login,
		Role:      params.Role,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt,
	}
	r.users[user.ID] = user
	r.creds[user.ID] = ud.Credentials{
		UserID:            user.ID,
		Email:             user.Email,
		Login:             user.Login,
		Role:              user.Role,
		PasswordHash:      params.PasswordHash,
		PasswordChangedAt: params.PasswordChangedAt,
	}
	r.byEmail[user.Email] = user.ID
	r.byLogin[user.Login] = user.ID
	return user, nil
}

func (r *memoryRepo) ChangeRole(_ context.Context, id ud.UserID, role ud.Role) (ud.User, error) {
	user, ok := r.users[id]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	user.Role = role
	r.users[id] = user
	credentials := r.creds[id]
	credentials.Role = role
	r.creds[id] = credentials
	return user, nil
}

func (r *memoryRepo) Update(_ context.Context, u *ud.User) (ud.User, error) {
	r.users[u.ID] = *u
	return *u, nil
}

func (r *memoryRepo) Delete(_ context.Context, u *ud.User) error {
	delete(r.users, u.ID)
	return nil
}

func (r *memoryRepo) GetByID(_ context.Context, id ud.UserID) (ud.User, error) {
	user, ok := r.users[id]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	return user, nil
}

func (r *memoryRepo) GetByEmail(_ context.Context, email ud.Email) (ud.User, error) {
	id, ok := r.byEmail[email]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	return r.GetByID(context.Background(), id)
}

func (r *memoryRepo) GetByLogin(_ context.Context, login ud.Login) (ud.User, error) {
	id, ok := r.byLogin[login]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	return r.GetByID(context.Background(), id)
}

func (r *memoryRepo) GetCredentials(_ context.Context, id ud.UserID) (ud.Credentials, error) {
	credentials, ok := r.creds[id]
	if !ok {
		return ud.Credentials{}, ud.ErrUserNotFound
	}
	return credentials, nil
}

func (r *memoryRepo) GetCredentialsByEmail(ctx context.Context, email ud.Email) (ud.Credentials, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return ud.Credentials{}, err
	}
	return r.GetCredentials(ctx, user.ID)
}

func (r *memoryRepo) GetCredentialsByLogin(ctx context.Context, login ud.Login) (ud.Credentials, error) {
	user, err := r.GetByLogin(ctx, login)
	if err != nil {
		return ud.Credentials{}, err
	}
	return r.GetCredentials(ctx, user.ID)
}

func (r *memoryRepo) CreateRefresh(_ context.Context, s ud.RefreshSession) error {
	r.sessions[s.TokenHash] = s
	return nil
}

func (r *memoryRepo) GetRefresh(_ context.Context, tokenHash string) (ud.RefreshSession, error) {
	session, ok := r.sessions[tokenHash]
	if !ok {
		return ud.RefreshSession{}, ud.ErrSessionNotFound
	}
	return session, nil
}

func (r *memoryRepo) RevokeRefresh(_ context.Context, tokenHash string) error {
	session, ok := r.sessions[tokenHash]
	if !ok {
		return ud.ErrSessionNotFound
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	r.sessions[tokenHash] = session
	return nil
}
