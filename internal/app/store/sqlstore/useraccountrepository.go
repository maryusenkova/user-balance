package sqlstore

import (
	"database/sql"
	"user_balance_microservice/internal/app/model"
)

type UserAccountRepository struct {
	store *Store
}

func (r *UserAccountRepository) Create(tx *sql.Tx, account *model.UserAccount) error {
	return tx.QueryRow(
		"INSERT INTO user_accounts (user_id, balance, reserved_balance) VALUES ($1, $2, $3) RETURNING user_id",
		account.User_id,
		account.Balance,
		0,
	).Scan(&account.User_id)
}

func (r *UserAccountRepository) Add(tx *sql.Tx, account *model.UserAccount) (*model.UserAccount, error) {
	if err := tx.QueryRow(
		"UPDATE user_accounts SET balance = balance + $1 where user_id = $2 RETURNING user_id, balance",
		account.Balance,
		account.User_id,
	).Scan(
		&account.User_id,
		&account.Balance,
	); err != nil {
		return nil, err
	}
	return account, nil
}

func (r *UserAccountRepository) Reserve(tx *sql.Tx, account *model.UserAccount) (*model.UserAccount, error) {
	if err := tx.QueryRow(
		"UPDATE user_accounts SET balance = balance - $1, reserved_balance = reserved_balance + $1 where user_id = $2 RETURNING user_id, balance",
		account.Balance,
		account.User_id,
	).Scan(
		&account.User_id,
		&account.Balance,
	); err != nil {
		return nil, err
	}
	return account, nil
}

func (r *UserAccountRepository) ConfirmReserve(tx *sql.Tx, account *model.UserAccount) (*model.UserAccount, error) {
	if err := tx.QueryRow(
		"UPDATE user_accounts SET reserved_balance = reserved_balance - $1 where user_id = $2 RETURNING user_id, balance",
		account.Balance,
		account.User_id,
	).Scan(
		&account.User_id,
		&account.Balance,
	); err != nil {
		return nil, err
	}
	return account, nil
}

func (r *UserAccountRepository) AbortReserve(tx *sql.Tx, account *model.UserAccount) (*model.UserAccount, error) {
	if err := tx.QueryRow(
		"UPDATE user_accounts SET balance = balance + $1, reserved_balance = reserved_balance - $1 where user_id = $2 RETURNING user_id, balance",
		account.Balance,
		account.User_id,
	).Scan(
		&account.User_id,
		&account.Balance,
	); err != nil {
		return nil, err
	}
	return account, nil
}

func (r *UserAccountRepository) FindById(id int) (*model.UserAccount, error) {
	account := &model.UserAccount{}
	if err := r.store.db.QueryRow(
		"SELECT user_id, balance from user_accounts where user_id=$1",
		id,
	).Scan(
		&account.User_id,
		&account.Balance,
	); err != nil {
		return nil, err
	}
	return account, nil
}
