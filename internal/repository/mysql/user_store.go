package mysql

import (
	"context"

	"btaskee-quiz/internal/model"

	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Save(ctx context.Context, user *model.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}

func (s *UserStore) Get(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.Get(ctx, username)
}
