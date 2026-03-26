package mysql

import (
	"context"

	"btaskee-quiz/internal/repository"

	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Save(ctx context.Context, user *repository.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}

func (s *UserStore) Get(ctx context.Context, uid string) (*repository.User, error) {
	var user repository.User
	err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
