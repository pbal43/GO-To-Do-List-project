package inmemory

import (
	"testing"
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"

	"github.com/stretchr/testify/assert"
)

func TestStorage_Users(t *testing.T) {
	storage := NewInMemoryStorage()

	user1 := usermodels.User{
		UUID:     "user1",
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "pass1",
	}

	user2 := usermodels.User{
		UUID:     "user2",
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
				savedUser := result.(usermodels.User)
				assert.Equal(t, "Alice", savedUser.Name)
				assert.Len(t, storage.users, 1)
			},
			expectError: nil,
		},
		{
			name: "SaveUser_duplicate_email",
			action: func() (any, error) {
				return storage.SaveUser(usermodels.User{
					UUID:     "user3",
					Name:     "Alice2",
					Email:    "alice@example.com",
					Password: "pass3",
				})
			},
			check:       func(_ *testing.T, _ any) {},
			expectError: usererrors.ErrUserIsAlreadyExist,
		},
		{
			name: "GetAllUsers_success",
			action: func() (any, error) {
				return storage.GetAllUsers()
			},
			check: func(t *testing.T, result any) {
				users := result.([]usermodels.User)
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
				user := result.(usermodels.User)
				assert.Equal(t, "Alice", user.Name)
			},
			expectError: nil,
		},
		{
			name: "GetUserByID_not_exist",
			action: func() (any, error) {
				return storage.GetUserByID("user404")
			},
			check:       func(_ *testing.T, _ any) {},
			expectError: usererrors.ErrUserNotExist,
		},
		{
			name: "GetUserByEmail_success",
			action: func() (any, error) {
				return storage.GetUserByEmail("alice@example.com")
			},
			check: func(t *testing.T, result any) {
				user := result.(usermodels.User)
				assert.Equal(t, "Alice", user.Name)
			},
			expectError: nil,
		},
		{
			name: "GetUserByEmail_not_exist",
			action: func() (any, error) {
				return storage.GetUserByEmail("unknown@example.com")
			},
			check:       func(_ *testing.T, _ any) {},
			expectError: usererrors.ErrUserNotExist,
		},
		{
			name: "UpdateUser_success",
			action: func() (any, error) {
				user1.Name = "Alice Updated"
				return storage.UpdateUser(user1)
			},
			check: func(t *testing.T, result any) {
				user := result.(usermodels.User)
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
			check:       func(_ *testing.T, _ any) {},
			expectError: usererrors.ErrUserNotExist,
		},
		{
			name: "DeleteUser_success",
			action: func() (any, error) {
				return nil, storage.DeleteUser("user1")
			},
			check: func(t *testing.T, _ any) {
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
			check:       func(_ *testing.T, _ any) {},
			expectError: usererrors.ErrUserNotExist,
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
