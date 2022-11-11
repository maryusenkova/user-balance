package apiserver

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"user_balance_microservice/internal/app/model"
	"user_balance_microservice/internal/app/store"
)

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
}

func newServer(store store.Store) *server {
	server := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		store:  store,
	}

	server.configureRouter()
	return server
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
func (s *server) configureRouter() {
	s.router.HandleFunc("/account/balance", s.getBalance()).Queries("id", "{[0-9]*?}").Methods("GET")
	s.router.HandleFunc("/account/add", s.handleBalanceAdd()).Methods("POST")
	s.router.HandleFunc("/reserve_money", s.handleReserveMoney()).Methods("POST")
	s.router.HandleFunc("/confirm_reserve", s.handleConfirm()).Methods("POST")
	s.router.HandleFunc("/abort_reserve", s.handleAbort()).Methods("POST")
}

func (s *server) getBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()
		user_id_str := v.Get("id")
		user_id, err := strconv.Atoi(user_id_str)
		if err != nil {
			s.respond(w, r, http.StatusBadRequest, map[string]string{"error": "User ID have to be a number"})
			return
		}
		account, err := s.store.UserAccount().FindById(user_id)
		if err != nil {
			err_str := fmt.Sprintf("No user with id = %s", user_id_str)
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": err_str})
			return
		}

		s.respond(w, r, http.StatusOK, account)
	}
}

func (s *server) handleBalanceAdd() http.HandlerFunc {
	type request struct {
		User_id int `json:"id"`
		Amount  int `json:"amount"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		tx, err := s.store.BeginTx()
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		created := true
		if _, err := s.store.UserAccount().FindById(req.User_id); err != nil {
			created = false
		}

		account := &model.UserAccount{
			User_id: req.User_id,
			Balance: req.Amount,
		}

		transaction := &model.Transaction{
			User_id:     req.User_id,
			Amount:      req.Amount,
			Description: "Пополнение счета",
			Closed_date: time.Now(),
			Success_flg: true,
			Type:        "add",
		}

		if !created {
			if err := s.store.UserAccount().Create(tx, account); err != nil {
				tx.Rollback()
				s.error(w, r, http.StatusUnprocessableEntity, err)
				return
			}
		} else {
			account, err = s.store.UserAccount().Add(tx, account)
			if err != nil {
				tx.Rollback()
				s.error(w, r, http.StatusUnprocessableEntity, err)
				return
			}
		}

		if err := s.store.Transaction().CreateAddTransaction(tx, transaction); err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		tx.Commit()
		s.respond(w, r, http.StatusOK, account)
	}
}

func (s *server) handleReserveMoney() http.HandlerFunc {
	type request struct {
		User_id    int `json:"id"`
		Service_id int `json:"serviceId"`
		Order_id   int `json:"orderId"`
		Amount     int `json:"amount"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		tx, err := s.store.BeginTx()
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		account, err := s.store.UserAccount().FindById(req.User_id)
		if err != nil {
			err_str := fmt.Sprintf("No user with id = %d", req.User_id)
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": err_str})
			return
		}
		if account.Balance < req.Amount {
			err_str := fmt.Sprintf("Not enough money for reserve. Current balance is %d", account.Balance)
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": err_str})
			return
		}
		reserve := &model.UserAccount{
			User_id: req.User_id,
			Balance: req.Amount,
		}

		reserve, err = s.store.UserAccount().Reserve(tx, reserve)
		if err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		transaction := &model.Transaction{
			User_id:     req.User_id,
			Amount:      req.Amount,
			Description: "Услуга",
			Service_id:  req.Service_id,
			Order_id:    req.Order_id,
			Type:        "reserve",
		}
		if err := s.store.Transaction().CreateReserveTransaction(tx, transaction); err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		tx.Commit()
		s.respond(w, r, http.StatusOK, reserve)
	}
}

func (s *server) handleConfirm() http.HandlerFunc {
	type request struct {
		User_id    int `json:"id"`
		Service_id int `json:"serviceId"`
		Order_id   int `json:"orderId"`
		Amount     int `json:"amount"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		tx, err := s.store.BeginTx()
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		transactionSearch := &model.Transaction{
			User_id:    req.User_id,
			Amount:     req.Amount,
			Service_id: req.Service_id,
			Order_id:   req.Order_id,
		}
		transaction, err := s.store.Transaction().GetTransaction(transactionSearch)
		if err != nil {
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": "No open reservation with such data"})
			return
		}

		reserve := &model.UserAccount{
			User_id: req.User_id,
			Balance: req.Amount,
		}

		reserve, err = s.store.UserAccount().ConfirmReserve(tx, reserve)
		if err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := s.store.Transaction().ConfirmReserveTransaction(tx, transaction.Id); err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		tx.Commit()
		s.respond(w, r, http.StatusOK, map[string]string{"success": "Money reserve confirmed"})
	}
}

func (s *server) handleAbort() http.HandlerFunc {
	type request struct {
		User_id    int `json:"id"`
		Service_id int `json:"serviceId"`
		Order_id   int `json:"orderId"`
		Amount     int `json:"amount"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		tx, err := s.store.BeginTx()
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		transactionSearch := &model.Transaction{
			User_id:    req.User_id,
			Amount:     req.Amount,
			Service_id: req.Service_id,
			Order_id:   req.Order_id,
		}
		transaction, err := s.store.Transaction().GetTransaction(transactionSearch)
		if err != nil {
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": "No open reservation with such data"})
			return
		}

		reserve := &model.UserAccount{
			User_id: req.User_id,
			Balance: req.Amount,
		}

		reserve, err = s.store.UserAccount().AbortReserve(tx, reserve)
		if err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := s.store.Transaction().AbortReserveTransaction(tx, transaction.Id); err != nil {
			tx.Rollback()
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		tx.Commit()
		s.respond(w, r, http.StatusOK, map[string]string{"success": "Money reserve aborted"})
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})

}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
