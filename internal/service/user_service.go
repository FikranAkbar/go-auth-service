package service

import "go-auth-service/internal/domain/user"

type UserService struct {
	userRepository user.RepositoryInterface
}

func NewUserService(userRepo user.RepositoryInterface) *UserService {
	return &UserService{userRepository: userRepo}
}

// Compile-time check to ensure UserService implements user.ServiceInterface
var _ user.ServiceInterface = (*UserService)(nil)
