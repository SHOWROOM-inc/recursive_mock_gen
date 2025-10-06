package usecases

import "github.com/SHOWROOM-inc/recursive_mock_gen/testing/samples/models"

type UserUseCase interface {
	GetUser(userID int64) (*models.User, error)
	Register(userID string, name string) error
}
