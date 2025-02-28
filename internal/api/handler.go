package api

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/w0ikid/megacalc/internal/service"
)

// Handler handles HTTP requests
type Handler struct {
	service *service.Service
}

// ExpressionRequest represents a request to calculate an expression
type ExpressionRequest struct {
	Expression string `json:"expression" binding:"required"`
}

// TaskResultRequest represents a request to set a task result
type TaskResultRequest struct {
	ID     string  `json:"id" binding:"required"`
	Result float64 `json:"result" binding:"required"`
}

// ExpressionResponse represents an expression response
type ExpressionResponse struct {
	ID     string             `json:"id"`
	Status service.ExpressionStatus `json:"status"`
	Result *float64           `json:"result,omitempty"`
}

// ExpressionsResponse represents a list of expressions
type ExpressionsResponse struct {
	Expressions []ExpressionResponse `json:"expressions"`
}

// ExpressionDetailResponse represents a single expression detail
type ExpressionDetailResponse struct {
	Expression ExpressionResponse `json:"expression"`
}

// TaskResponse represents a task response
type TaskResponse struct {
	Task *service.Task `json:"task,omitempty"`
}

// NewHandler creates a new API handler
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SetupRouter sets up the router
func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.Default()
	
	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api/v1")
	{
		api.POST("/calculate", h.CalculateExpression)
		api.GET("/expressions", h.GetExpressions)
		api.GET("/expressions/:id", h.GetExpression)
	}

	internal := r.Group("/internal")
	{
		internal.GET("/task", h.GetTask)
		internal.POST("/task", h.SetTaskResult)
	}

	// Serve static files for the web interface
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/index.html")

	return r
}

// CalculateExpression handles the request to calculate an expression
func (h *Handler) CalculateExpression(c *gin.Context) {
	var req ExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.SubmitExpression(req.Expression)
	if err != nil {
		log.Printf("Error submitting expression: %v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetExpressions handles the request to get all expressions
func (h *Handler) GetExpressions(c *gin.Context) {
	expressions := h.service.GetExpressions()
	
	var response ExpressionsResponse
	for _, expr := range expressions {
		response.Expressions = append(response.Expressions, ExpressionResponse{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetExpression handles the request to get an expression by ID
func (h *Handler) GetExpression(c *gin.Context) {
	id := c.Param("id")
	
	expr, found := h.service.GetExpression(id)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "expression not found"})
		return
	}

	c.JSON(http.StatusOK, ExpressionDetailResponse{
		Expression: ExpressionResponse{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		},
	})
}

// GetTask handles the request to get a task
func (h *Handler) GetTask(c *gin.Context) {
	task, found := h.service.GetTask()
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "no task available"})
		return
	}

	c.JSON(http.StatusOK, TaskResponse{Task: task})
}

// SetTaskResult handles the request to set a task result
func (h *Handler) SetTaskResult(c *gin.Context) {
	var req TaskResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err := h.service.SetTaskResult(req.ID, req.Result)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Start starts the server
func (h *Handler) Start(addr string) error {
	r := h.SetupRouter()
	return r.Run(addr)
}

// GetOperationTimes gets the operation times from environment variables
func GetOperationTimes() service.OperationTimes {
	return service.OperationTimes{
		Addition:       getEnvInt("TIME_ADDITION_MS", 1000),
		Subtraction:    getEnvInt("TIME_SUBTRACTION_MS", 1000),
		Multiplication: getEnvInt("TIME_MULTIPLICATIONS_MS", 1000),
		Division:       getEnvInt("TIME_DIVISIONS_MS", 1000),
	}
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}