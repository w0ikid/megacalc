package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/w0ikid/megacalc/internal/service"
)

func setupTestHandler() *Handler {
	svc := service.NewService(service.OperationTimes{
		Addition:       1,
		Subtraction:    1,
		Multiplication: 1,
		Division:       1,
	})
	return NewHandler(svc)
}

func TestCalculateExpression(t *testing.T) {
	h := setupTestHandler()
	router := h.SetupRouter()

	// Test valid expression
	reqBody := ExpressionRequest{Expression: "2+2"}
	jsonReq, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["id"])

	// Test invalid expression
	reqBody = ExpressionRequest{Expression: "2+*2"}
	jsonReq, _ = json.Marshal(reqBody)
	
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetExpressions(t *testing.T) {
	h := setupTestHandler()
	router := h.SetupRouter()
	
	// Add an expression
	reqBody := ExpressionRequest{Expression: "2+2"}
	jsonReq, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	// Get all expressions
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/expressions", nil)
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp ExpressionsResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Expressions, 1)
	assert.Equal(t, service.InProcess, resp.Expressions[0].Status)
}

func TestGetExpression(t *testing.T) {
	h := setupTestHandler()
	router := h.SetupRouter()
	
	// Add an expression
	reqBody := ExpressionRequest{Expression: "2+2"}
	jsonReq, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	var addResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &addResp)
	id := addResp["id"]
	
	// Get the expression
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/expressions/"+id, nil)
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp ExpressionDetailResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, id, resp.Expression.ID)
	assert.Equal(t, service.InProcess, resp.Expression.Status)
	
	// Get a non-existent expression
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/expressions/non-existent", nil)
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetTask(t *testing.T) {
	h := setupTestHandler()
	router := h.SetupRouter()
	
	// Add an expression to create tasks
	reqBody := ExpressionRequest{Expression: "2+2"}
	jsonReq, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	// Get a task
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/internal/task", nil)
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp TaskResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp.Task)
	assert.Equal(t, "2", resp.Task.Arg1)
	assert.Equal(t, "2", resp.Task.Arg2)
	assert.Equal(t, service.Addition, resp.Task.Operation)
}

func TestSetTaskResult(t *testing.T) {
	h := setupTestHandler()
	router := h.SetupRouter()
	
	// Add an expression to create tasks
	reqBody := ExpressionRequest{Expression: "2+2"}
	jsonReq, _ := json.Marshal(reqBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	// Get a task
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/internal/task", nil)
	
	router.ServeHTTP(w, req)
	
	var taskResp TaskResponse
	json.Unmarshal(w.Body.Bytes(), &taskResp)
	taskID := taskResp.Task.ID
	
	// Set the result
	resultReq := TaskResultRequest{
		ID:     taskID,
		Result: 4,
	}
	jsonReq, _ = json.Marshal(resultReq)
	
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Set result for non-existent task
	resultReq = TaskResultRequest{
		ID:     "non-existent",
		Result: 4,
	}
	jsonReq, _ = json.Marshal(resultReq)
	
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}