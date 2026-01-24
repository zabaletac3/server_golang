package auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/modules/users"
	"github.com/eren_dev/go_server/internal/shared/auth"
)

type Service struct {
	userRepo   users.UserRepository
	jwtService *auth.JWTService
}

func NewService(userRepo users.UserRepository, jwtService *auth.JWTService) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *Service) Register(ctx context.Context, dto *RegisterDTO) (*TokenResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.CreateWithPassword(ctx, dto.Name, dto.Email, string(hashedPassword))
	if err != nil {
		if err == users.ErrEmailExists {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *Service) Login(ctx context.Context, dto *LoginDTO) (*TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		if err == users.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	tokens, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
