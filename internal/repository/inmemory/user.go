package inmemory

import (
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"
)

// TODO: протестить локальное хранилище

func (storage *Storage) GetAllUsers() ([]usermodels.User, error) {
	var users []usermodels.User

	for _, user := range storage.users {
		users = append(users, user)
	}
	return users, nil
}

func (storage *Storage) SaveUser(user usermodels.User) (usermodels.User, error) {
	for _, userInMemory := range storage.users {
		if user.Email == userInMemory.Email {
			return usermodels.User{}, usererrors.ErrUserIsAlreadyExist
		}
	}

	storage.users[user.UUID] = user

	return user, nil
}

func (storage *Storage) GetUserByID(userID string) (usermodels.User, error) {
	user, ok := storage.users[userID]
	if !ok {
		return usermodels.User{}, usererrors.ErrUserNotExist
	}

	return user, nil
}

func (storage *Storage) GetUserByEmail(email string) (usermodels.User, error) {
	var user usermodels.User
	for _, userInMemory := range storage.users {
		if userInMemory.Email == email {
			user = userInMemory
			return user, nil
		}
	}
	return usermodels.User{}, usererrors.ErrUserNotExist
}

func (storage *Storage) UpdateUser(user usermodels.User) (usermodels.User, error) {
	for _, userInMemory := range storage.users {
		if userInMemory.UUID == user.UUID {
			userInMemory.Name = user.Name
			userInMemory.Email = user.Email
			userInMemory.Password = user.Password

			storage.users[user.UUID] = userInMemory
			return user, nil
		}
	}
	return usermodels.User{}, usererrors.ErrUserNotExist
}

func (storage *Storage) DeleteUser(userID string) error {
	_, ok := storage.users[userID]
	if !ok {
		return usererrors.ErrUserNotExist
	}
	delete(storage.users, userID)
	return nil
}
