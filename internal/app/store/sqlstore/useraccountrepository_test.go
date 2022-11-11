package sqlstore_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"user_balance_microservice/internal/app/model"
	"user_balance_microservice/internal/app/store/sqlstore"
)

func TestUserAccountRepository_Create(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("user_accounts")

	tx, err := db.Begin()
	assert.Nil(t, err)

	s := sqlstore.New(db)
	account := model.TestUserAccount(t)

	assert.NoError(t, s.UserAccount().Create(tx, account))
	assert.NotNil(t, account)
}

func TestUserAccountRepository_FindById(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("user_accounts")

	tx, err := db.Begin()

	s := sqlstore.New(db)

	id := 2
	_, err = s.UserAccount().FindById(id)
	assert.Error(t, err)

	s.UserAccount().Create(tx, &model.UserAccount{
		User_id: 3,
		Balance: 300,
	})
	account, err := s.UserAccount().FindById(3)
	assert.NoError(t, err)
	assert.NotNil(t, account)

}
