package apiserver

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"user_balance_microservice/internal/app/store/sqlstore"
)

func TestServer_handleBalanceAdd(t *testing.T) {
	config := GetConfig()
	db, err := newDB(config.DatabaseURL)
	assert.Nil(t, err)
	s := newServer(sqlstore.New(db))

	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]int{
				"id":     1,
				"amount": 100,
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/account/add", b)
			log.Println(req)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_handleReserveMoney(t *testing.T) {
	config := GetConfig()
	db, err := newDB(config.DatabaseURL)
	assert.Nil(t, err)
	s := newServer(sqlstore.New(db))

	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]int{
				"id":        1,
				"serviceId": 1,
				"orderId":   1234,
				"amount":    100,
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/reserve_money", b)
			log.Println(req)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestServer_handleConfirm(t *testing.T) {
	config := GetConfig()
	db, err := newDB(config.DatabaseURL)
	assert.Nil(t, err)
	s := newServer(sqlstore.New(db))

	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
	}{
		{
			name: "valid",
			payload: map[string]int{
				"id":        1,
				"serviceId": 1,
				"orderId":   1234,
				"amount":    100,
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/confirm_reserve", b)
			log.Println(req)
			s.ServeHTTP(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
