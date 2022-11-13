package sqlstore

import (
	"database/sql"
	"fmt"
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

func (r *TransactionRepository) GetMonthReport(month int, year int) (map[string]int, error) {
	var report map[string]int = make(map[string]int)
	rows, err := r.store.db.Query(
		`select s.name service, sum(amount) amount
				from transactions t
				join servicies s
					on t.service_id = s.id
				where t.success_flg = true
					and	extract(month from t.closed_date) = $1
					and extract(year from t.closed_date) = $2
				group by s.name`,
		month,
		year)
	if err != nil {
		return report, err
	}
	defer rows.Close()
	for rows.Next() {
		var service string
		var amount int
		if err := rows.Scan(&service, &amount); err != nil {
			return report, err
		}
		report[service] = amount
	}

	return report, err
}

func (r *TransactionRepository) GetAccountReport(userId int, orderCol, orderDir string, page, pageSize int) (*[]model.AccountTransaction, error) {
	report := []model.AccountTransaction{}
	query_str := fmt.Sprintf(`select 	amount, 
						description, 
						coalesce(order_id, 0) order_id, 
						coalesce(s.name, 'n/d') service,
						closed_date 
				from transactions t
				left join servicies s
				on t.service_id = s.id
				where success_flg=true
				and user_id = $1
				order by $2 %s
				offset $3 rows
				fetch next $4 rows only`, orderDir)
	rows, err := r.store.db.Query(query_str,
		userId,
		orderCol,
		(page-1)*pageSize,
		pageSize)
	if err != nil {
		return &report, err
	}
	defer rows.Close()
	for rows.Next() {
		record := model.AccountTransaction{}
		if err := rows.Scan(&record.Amount, &record.Description, &record.Order_id, &record.Service, &record.Closed_date); err != nil {
			return &report, err
		}
		report = append(report, record)
	}

	return &report, err
}
