package apiserver

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	s.router.HandleFunc("/get_report", s.handleGetReport()).Methods("POST")
	//s.router.HandleFunc("/csvreports/{filename}", s.getFile()).Methods("GET")
	s.router.HandleFunc("/account/history", s.handleGetHistory()).Methods("POST")
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
			Description: "Списание средств за услугу",
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

func (s *server) handleGetReport() http.HandlerFunc {
	type request struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		report, err := s.store.Transaction().GetMonthReport(req.Month, req.Year)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		if len(report) == 0 {
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": "No data for this month"})
			return
		}
		dir := "csvreports"
		_, err = os.Stat(dir)
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0777)
			if err != nil {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		path := fmt.Sprintf("%s/%d_%d_report.csv", dir, req.Month, req.Year)
		file, err := os.Create(path)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		defer file.Close()

		csvWriter := csv.NewWriter(file)
		defer csvWriter.Flush()
		for key, value := range report {
			row := []string{key, strconv.Itoa(value)}
			if err := csvWriter.Write(row); err != nil {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
		}
		s.respond(w, r, http.StatusOK, map[string]string{"file": fmt.Sprintf("./%s", file.Name())})
	}
}

func (s *server) handleGetHistory() http.HandlerFunc {
	type request struct {
		User_id   int    `json:"id"`
		Ordering  string `json:"ordering"`
		Page      int    `json:"page"`
		Page_size *int   `json:"pageSize,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		_, err := s.store.UserAccount().FindById(req.User_id)
		if err != nil {
			err_str := fmt.Sprintf("No user with id = %d", req.User_id)
			s.respond(w, r, http.StatusUnprocessableEntity, map[string]string{"error": err_str})
			return
		}
		orderDir := "ASC"
		if strings.HasPrefix(req.Ordering, "-") {
			orderDir = "DESC"
		}
		orderCol := "amount"
		if strings.Contains(req.Ordering, "date") {
			orderCol = "closed_date"
		}
		pageSize := 3
		if req.Page_size != nil {
			pageSize = *req.Page_size
		}
		report, err := s.store.Transaction().GetAccountReport(req.User_id, orderCol, orderDir, req.Page, pageSize)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, report)
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
