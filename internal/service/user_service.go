package service

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/model"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetByID(id string) (model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", id).Take(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *UserService) List(page, pageSize int, query string) ([]model.User, int64, error) {
	var (
		users []model.User
		total int64
	)

	dbQuery := s.db.Model(&model.User{})
	trimmed := strings.TrimSpace(query)
	if trimmed != "" {
		like := "%" + trimmed + "%"
		dbQuery = dbQuery.Where("username ILIKE ? OR display_name ILIKE ?", like, like)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if err := dbQuery.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

func (s *UserService) Create(req dto.CreateUserRequest) (model.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		DisplayName:  req.DisplayName,
		Email:        req.Email,
		Role:         req.Role,
		Status:       model.UserStatusActive,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return model.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *UserService) Update(userID string, req dto.UpdateUserRequest) (model.User, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return model.User{}, err
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.db.Save(&user).Error; err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *UserService) Disable(userID string) error {
	result := s.db.Model(&model.User{}).Where("id = ?", userID).Update("status", model.UserStatusDisabled)
	if result.Error != nil {
		return fmt.Errorf("disable user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *UserService) ResetPassword(userID, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	result := s.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", string(passwordHash))
	if result.Error != nil {
		return fmt.Errorf("reset user password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ToUserPayload(user model.User) dto.UserPayload {
	return dto.UserPayload{
		ID:          user.ID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
