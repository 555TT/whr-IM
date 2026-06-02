package service

import (
	"fmt"
	"strings"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
)

type GroupService struct {
	groupRepo  repository.GroupRepository
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
}

func NewGroupService(groupRepo repository.GroupRepository, friendRepo repository.FriendRepository, userRepo repository.UserRepository) *GroupService {
	return &GroupService{groupRepo: groupRepo, friendRepo: friendRepo, userRepo: userRepo}
}

type CreateGroupInput struct {
	Name      string
	MemberIDs []uint64
}

type GroupMemberItem struct {
	UserID             uint64 `json:"userId"`
	Username           string `json:"username"`
	Nickname           string `json:"nickname"`
	Avatar             string `json:"avatar"`
	PublicKey          string `json:"publicKey"`
	PublicKeyAlgorithm string `json:"publicKeyAlgorithm"`
}

type GroupDetail struct {
	ID      uint64            `json:"id"`
	Name    string            `json:"name"`
	OwnerID uint64            `json:"ownerId"`
	Members []GroupMemberItem `json:"members"`
}

type GroupListItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	OwnerID     uint64 `json:"ownerId"`
	MemberCount int    `json:"memberCount"`
}

const groupNameMaxLen = 50

func (s *GroupService) CreateGroup(userID uint64, input CreateGroupInput) (*GroupDetail, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if len([]rune(name)) > groupNameMaxLen {
		return nil, fmt.Errorf("group name too long")
	}
	if len(input.MemberIDs) == 0 {
		return nil, fmt.Errorf("at least one other member is required")
	}

	friendIDs, err := s.friendIDSet(userID)
	if err != nil {
		return nil, err
	}
	uniq := make(map[uint64]struct{}, len(input.MemberIDs)+1)
	uniq[userID] = struct{}{}
	for _, mid := range input.MemberIDs {
		if mid == userID {
			continue
		}
		if _, ok := friendIDs[mid]; !ok {
			return nil, fmt.Errorf("user %d is not your friend", mid)
		}
		uniq[mid] = struct{}{}
	}

	memberIDs := make([]uint64, 0, len(uniq))
	for uid := range uniq {
		memberIDs = append(memberIDs, uid)
	}

	group := &model.Group{Name: name, OwnerID: userID}
	if err := s.groupRepo.Create(group, memberIDs); err != nil {
		return nil, err
	}
	return s.buildGroupDetail(group)
}

func (s *GroupService) ListMyGroups(userID uint64) ([]GroupListItem, error) {
	groups, err := s.groupRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	items := make([]GroupListItem, 0, len(groups))
	for _, g := range groups {
		members, err := s.groupRepo.ListMembers(g.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, GroupListItem{
			ID:          g.ID,
			Name:        g.Name,
			OwnerID:     g.OwnerID,
			MemberCount: len(members),
		})
	}
	return items, nil
}

func (s *GroupService) GetGroupDetail(userID, groupID uint64) (*GroupDetail, error) {
	ok, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("you are not a member of this group")
	}
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}
	return s.buildGroupDetail(group)
}

func (s *GroupService) AddMembers(userID, groupID uint64, memberIDs []uint64) (*GroupDetail, error) {
	ok, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("you are not a member of this group")
	}
	if len(memberIDs) == 0 {
		return nil, fmt.Errorf("at least one member is required")
	}
	friendIDs, err := s.friendIDSet(userID)
	if err != nil {
		return nil, err
	}
	existingMembers, err := s.groupRepo.ListMembers(groupID)
	if err != nil {
		return nil, err
	}
	existing := make(map[uint64]struct{}, len(existingMembers))
	for _, m := range existingMembers {
		existing[m.UserID] = struct{}{}
	}

	toAdd := make([]uint64, 0, len(memberIDs))
	for _, mid := range memberIDs {
		if mid == userID {
			continue
		}
		if _, ok := existing[mid]; ok {
			continue
		}
		if _, ok := friendIDs[mid]; !ok {
			return nil, fmt.Errorf("user %d is not your friend", mid)
		}
		toAdd = append(toAdd, mid)
	}

	if err := s.groupRepo.AddMembers(groupID, toAdd); err != nil {
		return nil, err
	}
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}
	return s.buildGroupDetail(group)
}

func (s *GroupService) LeaveGroup(userID, groupID uint64) error {
	ok, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("you are not a member of this group")
	}
	return s.groupRepo.RemoveMember(groupID, userID)
}

func (s *GroupService) friendIDSet(userID uint64) (map[uint64]struct{}, error) {
	friends, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	set := make(map[uint64]struct{}, len(friends))
	for _, f := range friends {
		set[f.FriendID] = struct{}{}
	}
	return set, nil
}

func (s *GroupService) buildGroupDetail(group *model.Group) (*GroupDetail, error) {
	members, err := s.groupRepo.ListMembers(group.ID)
	if err != nil {
		return nil, err
	}
	items := make([]GroupMemberItem, 0, len(members))
	for _, m := range members {
		profile, err := s.userRepo.FindByID(m.UserID)
		if err != nil {
			return nil, err
		}
		items = append(items, GroupMemberItem{
			UserID:             profile.ID,
			Username:           profile.Username,
			Nickname:           profile.Nickname,
			Avatar:             profile.Avatar,
			PublicKey:          profile.PublicKey,
			PublicKeyAlgorithm: profile.PublicKeyAlgorithm,
		})
	}
	return &GroupDetail{
		ID:      group.ID,
		Name:    group.Name,
		OwnerID: group.OwnerID,
		Members: items,
	}, nil
}
