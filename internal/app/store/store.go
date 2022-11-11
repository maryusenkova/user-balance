package store

import "database/sql"

type Store interface {
	UserAccount() UserAccountRepository
	Transaction() TransactionRepository
	BeginTx() (*sql.Tx, error)
}
