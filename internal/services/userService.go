package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"gorm.io/gorm"
)

type UserService struct {
	userRepository UserStore
}

func NewUserService(userRepository UserStore) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) GetProfile(userId uuid.UUID) (*dto.UserResponseDto, error) {
	user, err := s.userRepository.GetUserById(userId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// The token is valid but the account no longer exists.
		return nil, exceptions.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	return dto.UserToResponseDto(user), nil
}

func (s *UserService) UpdateProfile(userId uuid.UUID, body dto.UpdateProfileDto) (*dto.UserResponseDto, error) {
	user, err := s.userRepository.GetUserById(userId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exceptions.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	if body.FirstName != nil {
		user.FirstName = strings.TrimSpace(*body.FirstName)
	}
	if body.LastName != nil {
		user.LastName = strings.TrimSpace(*body.LastName)
	}
	if body.InstagramNickname != nil {
		user.InstagramNickname = strings.TrimPrefix(strings.TrimSpace(*body.InstagramNickname), "@")
	}
	if body.BirthDate != nil {
		if *body.BirthDate == "" {
			user.BirthDate = nil
		} else {
			birthDate, err := time.Parse(dto.DateLayout, *body.BirthDate)
			if err != nil {
				return nil, exceptions.NewResponseError(
					exceptions.ErrUnprocessableEntity,
					fmt.Errorf("birth_date must use the %s format", dto.DateLayout),
				)
			}
			if birthDate.After(time.Now()) {
				return nil, exceptions.NewResponseError(exceptions.ErrUnprocessableEntity, errors.New("birth_date cannot be in the future"))
			}
			user.BirthDate = &birthDate
		}
	}

	if err := s.userRepository.UpdateUser(user); err != nil {
		return nil, err
	}
	return dto.UserToResponseDto(user), nil
}
