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
const defaultHomepageSkin = "aurora"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidPublicKey = errors.New("invalid public key update")
var ErrInvalidHomepageSkin = errors.New("invalid homepage skin")
var ErrProfileNotVisible = errors.New("profile is not visible to current user")

const supportedPublicKeyAlgorithm = "rsa-oaep-sha256"

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
	Nickname     string
	Gender       int
	Signature    string
	HomepageSkin string
}

type UpdatePublicKeyInput struct {
	PublicKey string
	Algorithm string
}

type PublicProfile struct {
	ID           uint64 `json:"id"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Signature    string `json:"signature"`
	HomepageSkin string `json:"homepageSkin"`
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
		Username:           input.Username,
		PasswordHash:       string(hash),
		Nickname:           input.Username,
		Avatar:             defaultAvatar,
		Gender:             0,
		Signature:          "",
		HomepageSkin:       defaultHomepageSkin,
		PublicKey:          "",
		PublicKeyAlgorithm: "",
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

func (s *AuthService) GetVisibleProfile(viewerID uint64, targetUserID uint64, friendRepo repository.FriendRepository) (*PublicProfile, error) {
	if viewerID != targetUserID {
		visible, err := friendRepo.AreFriends(viewerID, targetUserID)
		if err != nil {
			return nil, err
		}
		if !visible {
			return nil, ErrProfileNotVisible
		}
	}
	user, err := s.repo.FindByID(targetUserID)
	if err != nil {
		return nil, err
	}
	return &PublicProfile{
		ID:           user.ID,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Signature:    user.Signature,
		HomepageSkin: user.HomepageSkin,
	}, nil
}

func (s *AuthService) UpdateProfile(userID uint64, input UpdateProfileInput) (*model.User, error) {
	if !isAllowedHomepageSkin(input.HomepageSkin) {
		return nil, ErrInvalidHomepageSkin
	}
	return s.repo.UpdateProfile(userID, input.Nickname, input.Gender, input.Signature, input.HomepageSkin)
}

func (s *AuthService) UpdatePublicKey(userID uint64, input UpdatePublicKeyInput) (*model.User, error) {
	if input.PublicKey == "" {
		return nil, fmt.Errorf("%w: publicKey is required", ErrInvalidPublicKey)
	}
	if input.Algorithm == "" {
		return nil, fmt.Errorf("%w: algorithm is required", ErrInvalidPublicKey)
	}
	if input.Algorithm != supportedPublicKeyAlgorithm {
		return nil, fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidPublicKey, input.Algorithm)
	}

	return s.repo.UpdatePublicKey(userID, input.PublicKey, input.Algorithm)
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

func isAllowedHomepageSkin(skin string) bool {
	switch skin {
	case "aurora", "sunset", "galaxy", "mint", "peach":
		return true
	default:
		return false
	}
}
