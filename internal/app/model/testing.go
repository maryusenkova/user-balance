package model

import "testing"

func TestUserAccount(t *testing.T) *UserAccount {
	t.Helper()

	return &UserAccount{
		User_id: 1,
		Balance: 500,
	}
}
