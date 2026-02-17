package mobile_auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/eren_dev/go_server/internal/modules/owners"
	sharedAuth "github.com/eren_dev/go_server/internal/shared/auth"
)

type Service struct {
	ownerRepo  owners.OwnerRepository
	jwtService *sharedAuth.JWTService
}

func NewService(ownerRepo owners.OwnerRepository, jwtService *sharedAuth.JWTService) *Service {
	return &Service{
		ownerRepo:  ownerRepo,
		jwtService: jwtService,
	}
}

func (s *Service) Register(ctx context.Context, dto *RegisterDTO) (*TokenResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	owner, err := s.ownerRepo.Create(ctx, &owners.CreateOwnerDTO{
		Name:     dto.Name,
		Email:    dto.Email,
		Phone:    dto.Phone,
		Password: string(hashedPassword),
	})
	if err != nil {
		if err == owners.ErrEmailExists {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	tokens, err := s.jwtService.GenerateTokenPair(owner.ID.Hex(), owner.Email, sharedAuth.UserTypeOwner)
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
	owner, err := s.ownerRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		if err == owners.ErrOwnerNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(owner.Password), []byte(dto.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.jwtService.GenerateTokenPair(owner.ID.Hex(), owner.Email, sharedAuth.UserTypeOwner)
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

func (s *Service) GetOwnerInfo(ctx context.Context, ownerID string) (*OwnerInfo, error) {
	owner, err := s.ownerRepo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	return &OwnerInfo{
		ID:    owner.ID.Hex(),
		Name:  owner.Name,
		Email: owner.Email,
		Phone: owner.Phone,
	}, nil
}
