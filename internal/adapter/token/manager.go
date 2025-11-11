package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/nurpe/snowops-auth/internal/model"
)

type Claims struct {
	SessionID uuid.UUID      `json:"sid"`
	UserID    uuid.UUID      `json:"sub"`
	Role      model.UserRole `json:"role"`
	OrgID     uuid.UUID      `json:"org_id"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (m *Manager) GenerateAccessToken(user *model.User, sessionID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		SessionID: sessionID,
		UserID:    user.ID,
		Role:      user.Role,
		OrgID:     user.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
