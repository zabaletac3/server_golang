package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/eren_dev/go_server/internal/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type TokenType string
type UserType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

const (
	UserTypeStaff UserType = "staff"
	UserTypeOwner UserType = "owner"
)

type Claims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	UserType  UserType  `json:"user_type"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret            []byte
	expiration        time.Duration
	refreshExpiration time.Duration
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		secret:            []byte(cfg.JWTSecret),
		expiration:        cfg.JWTExpiration,
		refreshExpiration: cfg.JWTRefreshExpiration,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (s *JWTService) GenerateTokenPair(userID, email string, userType UserType) (*TokenPair, error) {
	accessToken, err := s.generateToken(userID, email, userType, AccessToken, s.expiration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, email, userType, RefreshToken, s.refreshExpiration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.expiration.Seconds()),
	}, nil
}

func (s *JWTService) generateToken(userID, email string, userType UserType, tokenType TokenType, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		UserType:  userType,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return nil, err
	}

	return s.GenerateTokenPair(claims.UserID, claims.Email, claims.UserType)
}
