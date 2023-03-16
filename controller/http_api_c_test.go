package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appditto/natrium-wallet-server/database"
	"github.com/appditto/natrium-wallet-server/net"
	"github.com/appditto/natrium-wallet-server/repository"
	"github.com/appditto/natrium-wallet-server/utils/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var app *chi.Mux
var controller *HttpController

func init() {
	// Mock http responses
	net.Client = &mocks.MockClient{}
	// Setup fiber
	mockDb, _ := database.NewConnection(&database.Config{
		Host:     os.Getenv("DB_MOCK_HOST"),
		Port:     os.Getenv("DB_MOCK_PORT"),
		Password: os.Getenv("DB_MOCK_PASS"),
		User:     os.Getenv("DB_MOCK_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   "testing",
	})
	database.DropAndCreateTables(mockDb)
	fcmRepo := &repository.FcmTokenRepo{
		DB: mockDb,
	}
	// Setup controllers
	rpcClient := net.RPCClient{
		Url: "http://localhost:8080",
	}
	controller = &HttpController{RPCClient: &rpcClient, BananoMode: false, FcmTokenRepo: fcmRepo, FcmClient: nil}
}

// Verify that unsupported actions are rejected
func TestUnsupportedAction(t *testing.T) {
	// Request JSON
	reqBody := map[string]interface{}{
		"action": "work_generate",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/api", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	controller.HandleAction(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "The requested action is not supported in this API", respJson["error"])
}

func TestSupportedAction(t *testing.T) {
	// Mock an account balance response
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: mocks.AccountBalanceResponse,
		}, nil
	}
	// Request JSON
	reqBody := map[string]interface{}{
		"action": "account_balance",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/api", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	controller.HandleAction(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "10000", respJson["balance"])
}
