package handler

import "go-auth-service/internal/domain/user"

type UserHandler struct {
	UserService user.ServiceInterface
}

func NewUserHandler(userService user.ServiceInterface) *UserHandler {
	return &UserHandler{UserService: userService}
}
