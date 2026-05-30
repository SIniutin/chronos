package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	audience      string
	now           func() time.Time
}

type TokenConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
	Audience      string
}

type AccessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenService(cfg TokenConfig) *TokenService {
	return &TokenService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		issuer:        cfg.Issuer,
		audience:      cfg.Audience,
		now:           time.Now,
	}
}

func (s *TokenService) NewPair(userID ud.UserID, role ud.Role) (ud.TokenPair, ud.RefreshSession, error) {
	now := s.now().UTC()
	accessExpires := now.Add(s.accessTTL)
	refreshExpires := now.Add(s.refreshTTL)

	accessToken, err := s.newAccessToken(userID, role, now, accessExpires)
	if err != nil {
		return ud.TokenPair{}, ud.RefreshSession{}, err
	}

	refreshToken, err := newOpaqueToken()
	if err != nil {
		return ud.TokenPair{}, ud.RefreshSession{}, err
	}

	sessionID := uuid.New()
	session := ud.RefreshSession{
		ID:        ud.SessionID(sessionID),
		UserID:    userID,
		TokenHash: s.HashRefresh(refreshToken),
		CreatedAt: now,
		ExpiresAt: refreshExpires,
	}

	return ud.TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessExpires:         accessExpires,
		RefreshTokenExpiresAt: refreshExpires,
	}, session, nil
}

func (s *TokenService) HashRefresh(token string) string {
	mac := hmac.New(sha256.New, s.refreshSecret)
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *TokenService) RefreshExpired(expiresAt time.Time) bool {
	return !expiresAt.After(s.now().UTC())
}

func (s *TokenService) newAccessToken(userID ud.UserID, role ud.Role, issuedAt, expiresAt time.Time) (string, error) {
	id := uuid.UUID(userID)
	claims := AccessClaims{
		Role: string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id.String(),
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func newOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
