package users_infra_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	users_security "github.com/SIniutin/history-app-backend/internal/modules/users/security"
	users_usecase "github.com/SIniutin/history-app-backend/internal/modules/users/usecase"
)

func TestHandlerAuthFlow(t *testing.T) {
	repo := newHandlerMemoryRepo()
	hasher := users_security.NewPasswordHasher()
	tokens := users_security.NewTokenService(users_security.TokenConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        "history-app-backend",
		Audience:      "history-app-users",
	})
	auth := users_infra.NewAuthMiddleware([]byte("access-secret"), "history-app-backend", "history-app-users")
	handler := users_infra.NewHandler(users_infra.Dependencies{
		Register:   users_usecase.NewRegisterUsecase(repo, hasher),
		Login:      users_usecase.NewLoginUsecase(repo, repo, hasher, tokens),
		Refresh:    users_usecase.NewRefreshUsecase(repo, repo, tokens),
		GetMe:      users_usecase.NewGetMeUsecase(repo),
		Logout:     users_usecase.NewLogoutUsecase(repo, tokens),
		ChangeRole: users_usecase.NewChangeUserRoleUsecase(repo),
		FindUser:   users_usecase.NewFindUserUsecase(repo),
		Auth:       auth,
	})
	router := handler.Routes()

	register := doJSON(router, http.MethodPost, "/auth/register", map[string]string{
		"email":    "user@example.com",
		"login":    "tester",
		"password": "password1",
	}, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d body=%s", register.Code, register.Body.String())
	}

	login := doJSON(router, http.MethodPost, "/auth/login", map[string]string{
		"identity": "tester",
		"password": "password1",
	}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", login.Code, login.Body.String())
	}
	var pair users_api.TokenPairResponse
	if err := json.NewDecoder(login.Body).Decode(&pair); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected token pair, got %+v", pair)
	}

	me := doJSON(router, http.MethodGet, "/users/me", nil, pair.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d body=%s", me.Code, me.Body.String())
	}

	refresh := doJSON(router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": pair.RefreshToken,
	}, "")
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status = %d body=%s", refresh.Code, refresh.Body.String())
	}
	var nextPair users_api.TokenPairResponse
	if err := json.NewDecoder(refresh.Body).Decode(&nextPair); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	if nextPair.RefreshToken == "" || nextPair.RefreshToken == pair.RefreshToken {
		t.Fatalf("expected rotated refresh token, got %+v", nextPair)
	}

	logout := doJSON(router, http.MethodPost, "/auth/logout", map[string]string{
		"refresh_token": nextPair.RefreshToken,
	}, nextPair.AccessToken)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d body=%s", logout.Code, logout.Body.String())
	}
}

func TestHandlerAdminLookupUser(t *testing.T) {
	repo := newHandlerMemoryRepo()
	hasher := users_security.NewPasswordHasher()
	tokens := users_security.NewTokenService(users_security.TokenConfig{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        "history-app-backend",
		Audience:      "history-app-users",
	})
	auth := users_infra.NewAuthMiddleware([]byte("access-secret"), "history-app-backend", "history-app-users")
	handler := users_infra.NewHandler(users_infra.Dependencies{
		Register:   users_usecase.NewRegisterUsecase(repo, hasher),
		Login:      users_usecase.NewLoginUsecase(repo, repo, hasher, tokens),
		GetMe:      users_usecase.NewGetMeUsecase(repo),
		ChangeRole: users_usecase.NewChangeUserRoleUsecase(repo),
		FindUser:   users_usecase.NewFindUserUsecase(repo),
		Auth:       auth,
	})
	router := handler.Routes()

	admin := createAndLoginForTest(t, router, "admin@example.com", "adminuser", "password1")
	target := createAndLoginForTest(t, router, "target@example.com", "targetuser", "password1")

	adminUser, err := repo.GetByLogin(context.Background(), ud.Login("adminuser"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ChangeRole(context.Background(), adminUser.ID, ud.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	admin = loginForTest(t, router, "adminuser", "password1")

	lookupByLogin := doJSON(router, http.MethodGet, "/admin/users/lookup?identity=targetuser", nil, admin.AccessToken)
	if lookupByLogin.Code != http.StatusOK {
		t.Fatalf("lookup by login status = %d body=%s", lookupByLogin.Code, lookupByLogin.Body.String())
	}
	var found users_api.AdminUserResponse
	if err := json.NewDecoder(lookupByLogin.Body).Decode(&found); err != nil {
		t.Fatalf("decode lookup: %v", err)
	}
	if found.ID == "" || found.Email != "target@example.com" || found.Login != "targetuser" {
		t.Fatalf("unexpected lookup response: %+v", found)
	}

	lookupByEmail := doJSON(router, http.MethodGet, "/admin/users/lookup?identity=target@example.com", nil, admin.AccessToken)
	if lookupByEmail.Code != http.StatusOK {
		t.Fatalf("lookup by email status = %d body=%s", lookupByEmail.Code, lookupByEmail.Body.String())
	}

	forbidden := doJSON(router, http.MethodGet, "/admin/users/lookup?identity=targetuser", nil, target.AccessToken)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("non-admin lookup status = %d body=%s", forbidden.Code, forbidden.Body.String())
	}
}

func TestHandlerProtectedRouteRejectsMissingToken(t *testing.T) {
	handler := users_infra.NewHandler(users_infra.Dependencies{
		Auth: users_infra.NewAuthMiddleware([]byte("access-secret"), "history-app-backend", "history-app-users"),
	})
	res := doJSON(handler.Routes(), http.MethodGet, "/users/me", nil, "")
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
}

func doJSON(handler http.Handler, method, path string, body any, accessToken string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func createAndLoginForTest(t *testing.T, router http.Handler, email, login, password string) users_api.TokenPairResponse {
	t.Helper()
	register := doJSON(router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email,
		"login":    login,
		"password": password,
	}, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register %s status = %d body=%s", login, register.Code, register.Body.String())
	}
	return loginForTest(t, router, login, password)
}

func loginForTest(t *testing.T, router http.Handler, identity, password string) users_api.TokenPairResponse {
	t.Helper()
	login := doJSON(router, http.MethodPost, "/auth/login", map[string]string{
		"identity": identity,
		"password": password,
	}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login %s status = %d body=%s", identity, login.Code, login.Body.String())
	}
	var pair users_api.TokenPairResponse
	if err := json.NewDecoder(login.Body).Decode(&pair); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	return pair
}

type handlerMemoryRepo struct {
	users    map[ud.UserID]ud.User
	creds    map[ud.UserID]ud.Credentials
	byEmail  map[ud.Email]ud.UserID
	byLogin  map[ud.Login]ud.UserID
	sessions map[string]ud.RefreshSession
}

func newHandlerMemoryRepo() *handlerMemoryRepo {
	return &handlerMemoryRepo{
		users:    make(map[ud.UserID]ud.User),
		creds:    make(map[ud.UserID]ud.Credentials),
		byEmail:  make(map[ud.Email]ud.UserID),
		byLogin:  make(map[ud.Login]ud.UserID),
		sessions: make(map[string]ud.RefreshSession),
	}
}

func (r *handlerMemoryRepo) Create(_ context.Context, params ud.CreateUserParams) (ud.User, error) {
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

func (r *handlerMemoryRepo) ChangeRole(_ context.Context, id ud.UserID, role ud.Role) (ud.User, error) {
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

func (r *handlerMemoryRepo) Update(_ context.Context, u *ud.User) (ud.User, error) {
	r.users[u.ID] = *u
	return *u, nil
}

func (r *handlerMemoryRepo) Delete(_ context.Context, u *ud.User) error {
	delete(r.users, u.ID)
	return nil
}

func (r *handlerMemoryRepo) GetByID(_ context.Context, id ud.UserID) (ud.User, error) {
	user, ok := r.users[id]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	return user, nil
}

func (r *handlerMemoryRepo) GetByEmail(_ context.Context, email ud.Email) (ud.User, error) {
	id, ok := r.byEmail[email]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	return r.users[id], nil
}

func (r *handlerMemoryRepo) GetByLogin(_ context.Context, login ud.Login) (ud.User, error) {
	id, ok := r.byLogin[login]
	if !ok {
		return ud.User{}, ud.ErrUserNotFound
	}
	return r.users[id], nil
}

func (r *handlerMemoryRepo) GetCredentials(_ context.Context, id ud.UserID) (ud.Credentials, error) {
	credentials, ok := r.creds[id]
	if !ok {
		return ud.Credentials{}, ud.ErrUserNotFound
	}
	return credentials, nil
}

func (r *handlerMemoryRepo) GetCredentialsByEmail(ctx context.Context, email ud.Email) (ud.Credentials, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return ud.Credentials{}, err
	}
	return r.GetCredentials(ctx, user.ID)
}

func (r *handlerMemoryRepo) GetCredentialsByLogin(ctx context.Context, login ud.Login) (ud.Credentials, error) {
	user, err := r.GetByLogin(ctx, login)
	if err != nil {
		return ud.Credentials{}, err
	}
	return r.GetCredentials(ctx, user.ID)
}

func (r *handlerMemoryRepo) CreateRefresh(_ context.Context, s ud.RefreshSession) error {
	r.sessions[s.TokenHash] = s
	return nil
}

func (r *handlerMemoryRepo) GetRefresh(_ context.Context, tokenHash string) (ud.RefreshSession, error) {
	session, ok := r.sessions[tokenHash]
	if !ok {
		return ud.RefreshSession{}, ud.ErrSessionNotFound
	}
	return session, nil
}

func (r *handlerMemoryRepo) RevokeRefresh(_ context.Context, tokenHash string) error {
	session, ok := r.sessions[tokenHash]
	if !ok {
		return ud.ErrSessionNotFound
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	r.sessions[tokenHash] = session
	return nil
}

var _ ud.UserRepository = (*handlerMemoryRepo)(nil)
var _ ud.SessionRepository = (*handlerMemoryRepo)(nil)
