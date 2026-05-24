package repository

import (
	"errors"
	"sync"

	"whr-im/server/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameExists = errors.New("username already exists")

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint64) (*model.User, error)
	UpdateProfile(userID uint64, nickname string, gender int, signature string) (*model.User, error)
}

type InMemoryUserRepository struct {
	mu         sync.RWMutex
	nextID     uint64
	users      map[uint64]*model.User
	userByName map[string]uint64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		nextID:     1,
		users:      make(map[uint64]*model.User),
		userByName: make(map[string]uint64),
	}
}

func (r *InMemoryUserRepository) Create(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.userByName[user.Username]; exists {
		return ErrUsernameExists
	}

	copyUser := *user
	copyUser.ID = r.nextID
	r.nextID++

	r.users[copyUser.ID] = &copyUser
	r.userByName[copyUser.Username] = copyUser.ID
	user.ID = copyUser.ID
	return nil
}

func (r *InMemoryUserRepository) FindByUsername(username string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.userByName[username]
	if !ok {
		return nil, ErrUserNotFound
	}

	user := *r.users[id]
	return &user, nil
}

func (r *InMemoryUserRepository) FindByID(id uint64) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}

	copyUser := *user
	return &copyUser, nil
}

func (r *InMemoryUserRepository) UpdateProfile(userID uint64, nickname string, gender int, signature string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	user.Nickname = nickname
	user.Gender = gender
	user.Signature = signature

	copyUser := *user
	return &copyUser, nil
}
