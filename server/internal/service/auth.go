package service

import (
	"errors"
	"fmt"
	"time"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const defaultAvatar = "https://api.dicebear.com/7.x/initials/svg?seed=default-user"

var ErrInvalidCredentials = errors.New("invalid credentials")

const minUsernameLength = 4
const maxUsernameLength = 20
const minPasswordLength = 6
const maxPasswordLength = 20

type AuthService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: []byte(jwtSecret)}
}

type RegisterInput struct {
	Username        string
	Password        string
	ConfirmPassword string
}

type UpdateProfileInput struct {
	Nickname  string
	Gender    int
	Signature string
}

func (s *AuthService) Register(input RegisterInput) (*model.User, error) {
	if err := validateCredentials(input.Username, input.Password); err != nil {
		return nil, err
	}
	if input.Password != input.ConfirmPassword {
		return nil, fmt.Errorf("passwords do not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     input.Username,
		PasswordHash: string(hash),
		Nickname:     input.Username,
		Avatar:       defaultAvatar,
		Gender:       0,
		Signature:    "",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	if err := validateCredentials(username, password); err != nil {
		return "", nil, err
	}

	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return signedToken, user, nil
}

func (s *AuthService) ParseToken(tokenString string) (uint64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidCredentials
	}

	value, ok := claims["userId"].(float64)
	if !ok {
		return 0, ErrInvalidCredentials
	}

	return uint64(value), nil
}

func (s *AuthService) GetProfile(userID uint64) (*model.User, error) {
	return s.repo.FindByID(userID)
}

func (s *AuthService) UpdateProfile(userID uint64, input UpdateProfileInput) (*model.User, error) {
	return s.repo.UpdateProfile(userID, input.Nickname, input.Gender, input.Signature)
}

func validateCredentials(username, password string) error {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return fmt.Errorf("username length must be between 4 and 20")
	}
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return fmt.Errorf("password length must be between 6 and 20")
	}
	return nil
}
