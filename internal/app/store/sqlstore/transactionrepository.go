package sqlstore

import (
	"database/sql"
	"user_balance_microservice/internal/app/model"
)

type TransactionRepository struct {
	store *Store
}

func (r *TransactionRepository) CreateReserveTransaction(tx *sql.Tx, transaction *model.Transaction) error {
	return tx.QueryRow(
		"INSERT INTO transactions (user_id, amount, description, order_id, service_id, type) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		transaction.User_id,
		transaction.Amount,
		transaction.Description,
		transaction.Order_id,
		transaction.Service_id,
		transaction.Type,
	).Scan(&transaction.Id)
}

func (r *TransactionRepository) CreateAddTransaction(tx *sql.Tx, transaction *model.Transaction) error {
	return tx.QueryRow(
		"INSERT INTO transactions (user_id, amount, description, closed_date, success_flg, type) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		transaction.User_id,
		transaction.Amount,
		transaction.Description,
		transaction.Closed_date,
		transaction.Success_flg,
		transaction.Type,
	).Scan(&transaction.Id)
}

func (r *TransactionRepository) GetTransaction(transaction *model.Transaction) (*model.Transaction, error) {
	if err := r.store.db.QueryRow(
		"select id from transactions where user_id = $1 and order_id=$2 and service_id=$3 and amount=$4 and closed_date is null",
		transaction.User_id,
		transaction.Order_id,
		transaction.Service_id,
		transaction.Amount,
	).Scan(&transaction.Id); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *TransactionRepository) ConfirmReserveTransaction(tx *sql.Tx, transactionId int) error {
	return tx.QueryRow(
		"update transactions set success_flg = true, closed_date = now() where id = $1 RETURNING id",
		transactionId,
	).Scan(&transactionId)
}

func (r *TransactionRepository) AbortReserveTransaction(tx *sql.Tx, transactionId int) error {
	return tx.QueryRow(
		"update transactions set closed_date = now() where id = $1 RETURNING id",
		transactionId,
	).Scan(&transactionId)
}
