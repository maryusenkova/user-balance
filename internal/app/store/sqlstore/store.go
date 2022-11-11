package sqlstore

import (
	"database/sql"
	_ "github.com/lib/pq"
	"user_balance_microservice/internal/app/store"
)

type Store struct {
	db                    *sql.DB
	userAccountRepository *UserAccountRepository
	transactionRepository *TransactionRepository
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) UserAccount() store.UserAccountRepository {
	if s.userAccountRepository != nil {
		return s.userAccountRepository
	}

	s.userAccountRepository = &UserAccountRepository{
		store: s,
	}
	return s.userAccountRepository
}

func (s *Store) Transaction() store.TransactionRepository {
	if s.transactionRepository != nil {
		return s.transactionRepository
	}

	s.transactionRepository = &TransactionRepository{
		store: s,
	}
	return s.transactionRepository
}

func (s *Store) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}
