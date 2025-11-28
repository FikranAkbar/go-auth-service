package service

import "go-auth-service/internal/repository"

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepository: userRepo}
}
