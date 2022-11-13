package store

import (
	"database/sql"
	"user_balance_microservice/internal/app/model"
)

type UserAccountRepository interface {
	Create(*sql.Tx, *model.UserAccount) error
	FindById(int) (*model.UserAccount, error)
	Add(*sql.Tx, *model.UserAccount) (*model.UserAccount, error)
	Reserve(*sql.Tx, *model.UserAccount) (*model.UserAccount, error)
	ConfirmReserve(*sql.Tx, *model.UserAccount) (*model.UserAccount, error)
	AbortReserve(*sql.Tx, *model.UserAccount) (*model.UserAccount, error)
	Transfer(*sql.Tx, int, int, int) (*model.UserAccount, error)
}

type TransactionRepository interface {
	CreateReserveTransaction(*sql.Tx, *model.Transaction) error
	CreateAddTransaction(*sql.Tx, *model.Transaction) error
	GetTransaction(*model.Transaction) (*model.Transaction, error)
	ConfirmReserveTransaction(*sql.Tx, int) error
	AbortReserveTransaction(*sql.Tx, int) error
	GetMonthReport(int, int) (map[string]int, error)
	GetAccountReport(int, string, string, int, int) (*[]model.AccountTransaction, error)
}
