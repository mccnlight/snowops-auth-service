package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/nurpe/snowops-auth/internal/adapter/password"
	"github.com/nurpe/snowops-auth/internal/adapter/sms"
	"github.com/nurpe/snowops-auth/internal/adapter/token"
	"github.com/nurpe/snowops-auth/internal/config"
	"github.com/nurpe/snowops-auth/internal/model"
	"github.com/nurpe/snowops-auth/internal/repository"
)

type AuthService struct {
	users         repository.UserRepository
	organizations repository.OrganizationRepository
	sessions      repository.UserSessionRepository
	smsCodes      repository.SmsCodeRepository
	password      password.Hasher
	smsSender     sms.Sender
	tokens        *token.Manager
	cfg           *config.Config
}

type AuthMeta struct {
	UserAgent string
	ClientIP  string
}

type AuthResult struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Organization   string         `json:"organization"`
	Role           model.UserRole `json:"role"`
	Phone          *string        `json:"phone,omitempty"`
	Login          *string        `json:"login,omitempty"`
	IsNew          bool           `json:"is_new,omitempty"`
}

func NewAuthService(
	users repository.UserRepository,
	organizations repository.OrganizationRepository,
	sessions repository.UserSessionRepository,
	smsCodes repository.SmsCodeRepository,
	password password.Hasher,
	smsSender sms.Sender,
	tokens *token.Manager,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		users:         users,
		organizations: organizations,
		sessions:      sessions,
		smsCodes:      smsCodes,
		password:      password,
		smsSender:     smsSender,
		tokens:        tokens,
		cfg:           cfg,
	}
}

func (s *AuthService) Login(ctx context.Context, login, pass string, meta AuthMeta) (*AuthResult, error) {
	user, err := s.users.FindByLogin(ctx, login)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	if err := s.password.Compare(*user.PasswordHash, pass); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) SendCode(ctx context.Context, phone string) (string, error) {
	now := time.Now()
	if s.cfg.SMS.DailySendLimit > 0 {
		from := now.Add(-24 * time.Hour)
		count, err := s.smsCodes.CountActiveInRange(ctx, phone, from)
		if err != nil {
			return "", err
		}
		if count >= int64(s.cfg.SMS.DailySendLimit) {
			return "", fmt.Errorf("daily sms limit reached")
		}
	}

	if _, err := s.users.FindByPhone(ctx, phone); errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrUserNotFound
	} else if err != nil {
		return "", err
	}

	code := generateNumericCode(s.cfg.SMS.CodeLength)
	smsCode := &model.SmsCode{
		Phone:     phone,
		Code:      code,
		ExpiresAt: now.Add(s.cfg.SMS.CodeTTL),
	}

	if err := s.smsCodes.Create(ctx, smsCode); err != nil {
		return "", err
	}

	if err := s.smsSender.Send(ctx, phone, sms.FormatAuthCode(code)); err != nil {
		return "", err
	}

	return maskPhone(phone), nil
}

func (s *AuthService) VerifyCode(ctx context.Context, phone, code string, meta AuthMeta) (*AuthResult, error) {
	smsCode, err := s.smsCodes.FindLatest(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCodeInvalid
	}
	if err != nil {
		return nil, err
	}

	if smsCode.IsUsed {
		return nil, ErrCodeInvalid
	}
	if smsCode.ExpiresAt.Before(time.Now()) {
		return nil, ErrCodeExpired
	}
	if smsCode.Code != code {
		return nil, ErrCodeInvalid
	}

	if err := s.smsCodes.MarkUsed(ctx, smsCode.ID.String()); err != nil {
		return nil, err
	}

	user, err := s.users.FindByPhone(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	result, err := s.createSessionAndTokens(ctx, user, meta)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta AuthMeta) (*AuthResult, error) {
	hash := hashToken(refreshToken)

	session, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionRevoked
	}

	user, err := s.users.FindByID(ctx, session.UserID.String())
	if err != nil {
		return nil, err
	}

	if err := s.sessions.Revoke(ctx, session.ID.String()); err != nil {
		return nil, err
	}

	return s.createSessionAndTokens(ctx, user, meta)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	session, err := s.sessions.FindByTokenHash(ctx, hash)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrSessionNotFound
	}
	if err != nil {
		return err
	}
	return s.sessions.Revoke(ctx, session.ID.String())
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*UserInfo, error) {
	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	info := toUserInfo(user)
	return &info, nil
}

func (s *AuthService) createSessionAndTokens(ctx context.Context, user *model.User, meta AuthMeta) (*AuthResult, error) {
	sessionID := uuid.New()
	refreshToken := uuid.NewString()
	refreshHash := hashToken(refreshToken)

	session := &model.UserSession{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        time.Now().Add(s.cfg.JWT.RefreshTTL),
		UserAgent:        meta.UserAgent,
		ClientIP:         meta.ClientIP,
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := s.tokens.GenerateAccessToken(user, session.ID, s.cfg.JWT.AccessTTL)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserInfo(user),
	}, nil
}

func generateNumericCode(length int) string {
	if length <= 0 {
		length = 6
	}
	const digits = "0123456789"
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	last4 := phone[len(phone)-4:]
	prefix := phonePrefix(phone)
	return fmt.Sprintf("%s***%s", prefix, last4)
}

func phonePrefix(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	if len(phone) <= 7 {
		return phone[:len(phone)-4]
	}
	return phone[:len(phone)-7]
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func toUserInfo(user *model.User) UserInfo {
	var orgName string
	if user.Organization.ID != uuid.Nil {
		orgName = user.Organization.Name
	}
	return UserInfo{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		Organization:   orgName,
		Role:           user.Role,
		Phone:          user.Phone,
		Login:          user.Login,
	}
}
