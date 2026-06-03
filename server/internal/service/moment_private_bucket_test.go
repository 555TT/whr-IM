package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"whr-im/server/internal/model"
)

type momentRepoStub struct {
	created *model.Moment
}

func (s *momentRepoStub) Create(moment *model.Moment) error {
	moment.ID = 1
	moment.CreatedAt = time.Now()
	s.created = moment
	return nil
}
func (s *momentRepoStub) ListVisibleForUser(userID uint64, friendIDs []uint64) ([]model.Moment, error) {
	return []model.Moment{*s.created}, nil
}
func (s *momentRepoStub) ListByUserID(userID uint64) ([]model.Moment, error) {
	return []model.Moment{*s.created}, nil
}
func (s *momentRepoStub) FindByID(id uint64) (*model.Moment, error)             { return s.created, nil }
func (s *momentRepoStub) Delete(momentID uint64) error                          { return nil }
func (s *momentRepoStub) Like(momentID uint64, userID uint64) error             { return nil }
func (s *momentRepoStub) Unlike(momentID uint64, userID uint64) error           { return nil }
func (s *momentRepoStub) CountLikes(momentID uint64) (int64, error)             { return 0, nil }
func (s *momentRepoStub) HasLiked(momentID uint64, userID uint64) (bool, error) { return false, nil }
func (s *momentRepoStub) CreateComment(comment *model.MomentComment) error      { return nil }
func (s *momentRepoStub) ListComments(momentID uint64) ([]model.MomentComment, error) {
	return nil, nil
}

type friendRepoStub struct{}

func (s *friendRepoStub) CreateRequest(request *model.FriendRequest) error { return nil }
func (s *friendRepoStub) ListIncomingRequests(userID uint64) ([]model.FriendRequest, error) {
	return nil, nil
}
func (s *friendRepoStub) HandleRequest(requestID uint64, userID uint64, status string) (*model.FriendRequest, error) {
	return nil, nil
}
func (s *friendRepoStub) CreateFriendPair(userID uint64, friendID uint64) error { return nil }
func (s *friendRepoStub) ListFriends(userID uint64) ([]model.Friend, error)     { return nil, nil }
func (s *friendRepoStub) AreFriends(userID uint64, friendID uint64) (bool, error) {
	return true, nil
}

type userRepoStub struct{}

func (s *userRepoStub) Create(user *model.User) error                       { return nil }
func (s *userRepoStub) FindByUsername(username string) (*model.User, error) { return nil, nil }
func (s *userRepoStub) FindByID(id uint64) (*model.User, error) {
	return &model.User{ID: id, Nickname: "alice", Avatar: "avatar"}, nil
}
func (s *userRepoStub) UpdateProfile(userID uint64, nickname string, gender int, signature string, homepageSkin string) (*model.User, error) {
	return nil, nil
}
func (s *userRepoStub) UpdatePublicKey(userID uint64, publicKey string, algorithm string) (*model.User, error) {
	return nil, nil
}

type objectStorageStub struct{}

func (s *objectStorageStub) UploadImage(ctx context.Context, objectKey string, contentType string, data []byte) (string, error) {
	return "http://upload.example/" + objectKey, nil
}
func (s *objectStorageStub) ObjectURL(ctx context.Context, objectKey string) (string, error) {
	return "http://signed.example/" + objectKey + "?token=abc", nil
}

func TestMomentServiceStoresObjectKeysButReturnsReadableURLs(t *testing.T) {
	repo := &momentRepoStub{}
	service := NewMomentService(repo, &friendRepoStub{}, &userRepoStub{}, &objectStorageStub{})

	_, err := service.Create(1, CreateMomentInput{
		Content:   "hello",
		ImageKeys: []string{"uploads/users/1/demo.png"},
	})
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if repo.created == nil {
		t.Fatal("expected moment to be created")
	}
	if !strings.Contains(repo.created.ImagesJSON, "uploads/users/1/demo.png") {
		t.Fatalf("expected stored image key, got %q", repo.created.ImagesJSON)
	}

	moments, err := service.ListVisible(1)
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(moments) != 1 {
		t.Fatalf("expected 1 moment, got %d", len(moments))
	}
	if len(moments[0].Images) != 1 {
		t.Fatalf("expected 1 image in response, got %d", len(moments[0].Images))
	}
	if moments[0].Images[0] != "http://signed.example/uploads/users/1/demo.png?token=abc" {
		t.Fatalf("expected signed image url, got %q", moments[0].Images[0])
	}
}
