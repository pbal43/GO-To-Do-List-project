package inmemory

import (
	"testing"
	"toDoList/internal/domain/user/user_errors"
	"toDoList/internal/domain/user/user_models"

	"github.com/stretchr/testify/assert"
)

func TestStorage_Users(t *testing.T) {
	storage := NewInMemoryStorage()

	user1 := user_models.User{
		Uuid:     "user1",
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "pass1",
	}

	user2 := user_models.User{
		Uuid:     "user2",
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "pass2",
	}

	tests := []struct {
		name        string
		action      func() (any, error)
		check       func(t *testing.T, result any)
		expectError error
	}{
		{
			name: "SaveUser_success",
			action: func() (any, error) {
				return storage.SaveUser(user1)
			},
			check: func(t *testing.T, result any) {
				savedUser := result.(user_models.User)
				assert.Equal(t, "Alice", savedUser.Name)
				assert.Len(t, storage.users, 1)
			},
			expectError: nil,
		},
		{
			name: "SaveUser_duplicate_email",
			action: func() (any, error) {
				return storage.SaveUser(user_models.User{
					Uuid:     "user3",
					Name:     "Alice2",
					Email:    "alice@example.com",
					Password: "pass3",
				})
			},
			check:       func(t *testing.T, result any) {},
			expectError: user_errors.ErrorUserIsAlreadyExist,
		},
		{
			name: "GetAllUsers_success",
			action: func() (any, error) {
				return storage.GetAllUsers()
			},
			check: func(t *testing.T, result any) {
				users := result.([]user_models.User)
				assert.Len(t, users, 1)
				assert.Equal(t, "Alice", users[0].Name)
			},
			expectError: nil,
		},
		{
			name: "GetUserByID_success",
			action: func() (any, error) {
				return storage.GetUserByID("user1")
			},
			check: func(t *testing.T, result any) {
				user := result.(user_models.User)
				assert.Equal(t, "Alice", user.Name)
			},
			expectError: nil,
		},
		{
			name: "GetUserByID_not_exist",
			action: func() (any, error) {
				return storage.GetUserByID("user404")
			},
			check:       func(t *testing.T, result any) {},
			expectError: user_errors.ErrorUserNotExist,
		},
		{
			name: "GetUserByEmail_success",
			action: func() (any, error) {
				return storage.GetUserByEmail("alice@example.com")
			},
			check: func(t *testing.T, result any) {
				user := result.(user_models.User)
				assert.Equal(t, "Alice", user.Name)
			},
			expectError: nil,
		},
		{
			name: "GetUserByEmail_not_exist",
			action: func() (any, error) {
				return storage.GetUserByEmail("unknown@example.com")
			},
			check:       func(t *testing.T, result any) {},
			expectError: user_errors.ErrorUserNotExist,
		},
		{
			name: "UpdateUser_success",
			action: func() (any, error) {
				user1.Name = "Alice Updated"
				return storage.UpdateUser(user1)
			},
			check: func(t *testing.T, result any) {
				user := result.(user_models.User)
				assert.Equal(t, "Alice Updated", user.Name)
				assert.Equal(t, "Alice Updated", storage.users["user1"].Name)
			},
			expectError: nil,
		},
		{
			name: "UpdateUser_not_exist",
			action: func() (any, error) {
				return storage.UpdateUser(user2)
			},
			check:       func(t *testing.T, result any) {},
			expectError: user_errors.ErrorUserNotExist,
		},
		{
			name: "DeleteUser_success",
			action: func() (any, error) {
				return nil, storage.DeleteUser("user1")
			},
			check: func(t *testing.T, result any) {
				_, ok := storage.users["user1"]
				assert.False(t, ok)
			},
			expectError: nil,
		},
		{
			name: "DeleteUser_not_exist",
			action: func() (any, error) {
				return nil, storage.DeleteUser("user404")
			},
			check:       func(t *testing.T, result any) {},
			expectError: user_errors.ErrorUserNotExist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.action()
			if tc.expectError != nil {
				assert.ErrorIs(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			tc.check(t, result)
		})
	}
}
