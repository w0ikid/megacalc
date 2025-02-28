package service

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Operation represents a mathematical operation
type Operation string

const (
	Addition       Operation = "+"
	Subtraction    Operation = "-"
	Multiplication Operation = "*"
	Division       Operation = "/"
)

// ExpressionStatus represents the status of an expression evaluation
type ExpressionStatus string

const (
	Pending   ExpressionStatus = "pending"
	InProcess ExpressionStatus = "in_process"
	Completed ExpressionStatus = "completed"
	Failed    ExpressionStatus = "failed"
)

// ExpressionData represents an expression with its evaluation status
type ExpressionData struct {
	ID         string           `json:"id"`
	Expression string           `json:"expression"`
	Status     ExpressionStatus `json:"status"`
	Result     *float64         `json:"result,omitempty"`
}

// Task represents a computational task
type Task struct {
	ID            string    `json:"id"`
	ExpressionID  string    `json:"expression_id"`
	Arg1          string    `json:"arg1"`
	Arg2          string    `json:"arg2"`
	Operation     Operation `json:"operation"`
	OperationTime int       `json:"operation_time"`
	Result        *float64  `json:"result,omitempty"`
	Status        string    `json:"status"`
	Dependencies  []string  `json:"-"`
}

// OperationTimes holds the configured durations for each operation
type OperationTimes struct {
	Addition       int
	Subtraction    int
	Multiplication int
	Division       int
}

// Service handles the business logic of the calculator
type Service struct {
	expressions      map[string]*ExpressionData
	tasks            map[string]*Task
	taskQueue        []string
	completedTasks   map[string]bool
	readyTasks       map[string]bool
	opTimes          OperationTimes
	mu               sync.RWMutex
	taskIDCounter    int
	dependencyGraph  map[string][]string
	reverseDependencies map[string][]string
}

// NewService creates a new calculator service
func NewService(opTimes OperationTimes) *Service {
	return &Service{
		expressions:      make(map[string]*ExpressionData),
		tasks:            make(map[string]*Task),
		taskQueue:        []string{},
		completedTasks:   make(map[string]bool),
		readyTasks:       make(map[string]bool),
		opTimes:          opTimes,
		taskIDCounter:    0,
		dependencyGraph:  make(map[string][]string),
		reverseDependencies: make(map[string][]string),
	}
}

// SubmitExpression adds a new expression to be calculated
func (s *Service) SubmitExpression(expression string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean the expression by removing spaces
	expression = strings.ReplaceAll(expression, " ", "")

	// Create a new expression entry
	id := uuid.New().String()
	expr := &ExpressionData{
		ID:         id,
		Expression: expression,
		Status:     Pending,
	}
	s.expressions[id] = expr

	// Parse the expression and create tasks
	err := s.parseExpression(id, expression)
	if err != nil {
		expr.Status = Failed
		return "", err
	}

	expr.Status = InProcess
	return id, nil
}

// GetExpressions returns all expressions
func (s *Service) GetExpressions() []ExpressionData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []ExpressionData
	for _, expr := range s.expressions {
		result = append(result, *expr)
	}
	return result
}

// GetExpression returns an expression by its ID
func (s *Service) GetExpression(id string) (*ExpressionData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expr, ok := s.expressions[id]
	if !ok {
		return nil, false
	}
	return expr, true
}

// GetTask returns the next task to be processed
func (s *Service) GetTask() (*Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find a ready task
	for taskID := range s.readyTasks {
		task := s.tasks[taskID]
		task.Status = "processing"
		delete(s.readyTasks, taskID)
		return task, true
	}

	return nil, false
}

// SetTaskResult sets the result of a task
func (s *Service) SetTaskResult(id string, result float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	// Set the result
	task.Result = &result
	s.completedTasks[id] = true
	
	// Update the expression if this was the final task
	s.updateExpressionStatus(task.ExpressionID)
	
	// Update dependencies
	s.updateDependencies(id, result)

	return nil
}

// updateDependencies updates the dependent tasks and adds them to the ready queue if all dependencies are met
func (s *Service) updateDependencies(taskID string, result float64) {
	for _, depID := range s.reverseDependencies[taskID] {
		depTask := s.tasks[depID]
		
		// Update the argument with the result
		if depTask.Arg1 == taskID {
			depTask.Arg1 = fmt.Sprintf("%f", result)
		}
		if depTask.Arg2 == taskID {
			depTask.Arg2 = fmt.Sprintf("%f", result)
		}
		
		// Check if all dependencies are completed
		allDepsCompleted := true
		for _, dID := range depTask.Dependencies {
			if !s.completedTasks[dID] {
				allDepsCompleted = false
				break
			}
		}
		
		// If all dependencies are completed, add to ready tasks
		if allDepsCompleted {
			s.readyTasks[depID] = true
		}
	}
}

// updateExpressionStatus checks if all tasks for an expression are completed and updates the status
func (s *Service) updateExpressionStatus(exprID string) {
	expr := s.expressions[exprID]
	
	// Check if all tasks are completed
	allCompleted := true
	var finalResult *float64
	
	for taskID, task := range s.tasks {
		if task.ExpressionID == exprID {
			if task.Result == nil {
				allCompleted = false
				break
			}
			// If this is the root task (no dependencies on it), it's the final result
			if len(s.reverseDependencies[taskID]) == 0 {
				finalResult = task.Result
			}
		}
	}
	
	if allCompleted && finalResult != nil {
		expr.Status = Completed
		expr.Result = finalResult
	}
}

// parseExpression parses the expression and creates tasks
func (s *Service) parseExpression(exprID, expression string) error {
	// We'll implement a simple parsing algorithm for expressions
	// This parser handles basic operations and respects operator precedence
	
	// First, we'll tokenize the expression
	tokens, err := tokenize(expression)
	if err != nil {
		return err
	}
	
	// Apply the shunting yard algorithm to handle operator precedence
	output, err := shuntingYard(tokens)
	if err != nil {
		return err
	}
	
	// Create tasks from the postfix notation
	return s.createTasksFromPostfix(exprID, output)
}

// createTasksFromPostfix creates tasks from postfix notation
func (s *Service) createTasksFromPostfix(exprID string, postfix []string) error {
	var stack []string
	
	for _, token := range postfix {
		if isOperator(token) {
			// Pop the top two values from the stack
			if len(stack) < 2 {
				return fmt.Errorf("invalid expression: not enough operands for operator %s", token)
			}
			
			arg2 := stack[len(stack)-1]
			arg1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			
			// Create a new task
			taskID := s.createTask(exprID, arg1, arg2, Operation(token))
			
			// Push the task ID back onto the stack
			stack = append(stack, taskID)
		} else {
			// Push the operand onto the stack
			stack = append(stack, token)
		}
	}
	
	// After processing, there should be exactly one item on the stack (the final result)
	if len(stack) != 1 {
		return fmt.Errorf("invalid expression: too many values left on stack")
	}
	
	// Find tasks with no dependencies and mark them as ready
	for taskID, task := range s.tasks {
		if task.ExpressionID == exprID {
			// If both arguments are not task IDs, this task is ready
			_, isArg1TaskID := s.tasks[task.Arg1]
			_, isArg2TaskID := s.tasks[task.Arg2]
			
			if !isArg1TaskID && !isArg2TaskID {
				s.readyTasks[taskID] = true
			}
		}
	}
	
	return nil
}

// createTask creates a new task and adds it to the service
func (s *Service) createTask(exprID, arg1, arg2 string, operation Operation) string {
	s.taskIDCounter++
	taskID := fmt.Sprintf("task_%d", s.taskIDCounter)
	
	// Determine operation time
	opTime := 0
	switch operation {
	case Addition:
		opTime = s.opTimes.Addition
	case Subtraction:
		opTime = s.opTimes.Subtraction
	case Multiplication:
		opTime = s.opTimes.Multiplication
	case Division:
		opTime = s.opTimes.Division
	}
	
	// Create the task
	task := &Task{
		ID:            taskID,
		ExpressionID:  exprID,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     operation,
		OperationTime: opTime,
		Status:        "pending",
		Dependencies:  []string{},
	}
	
	// Set up dependencies
	if _, isTaskID := s.tasks[arg1]; isTaskID {
		task.Dependencies = append(task.Dependencies, arg1)
		
		// Add this task to reverse dependencies
		s.reverseDependencies[arg1] = append(s.reverseDependencies[arg1], taskID)
	}
	
	if _, isTaskID := s.tasks[arg2]; isTaskID {
		task.Dependencies = append(task.Dependencies, arg2)
		
		// Add this task to reverse dependencies
		s.reverseDependencies[arg2] = append(s.reverseDependencies[arg2], taskID)
	}
	
	s.tasks[taskID] = task
	s.dependencyGraph[taskID] = task.Dependencies
	
	return taskID
}

// Helper functions

// tokenize converts a string expression into tokens
func tokenize(expression string) ([]string, error) {
	var tokens []string
	var currentNumber string
	
	for i := 0; i < len(expression); i++ {
		char := string(expression[i])
		
		switch {
		case char >= "0" && char <= "9" || char == ".":
			currentNumber += char
		case isOperator(char):
			if currentNumber != "" {
				tokens = append(tokens, currentNumber)
				currentNumber = ""
			}
			tokens = append(tokens, char)
		case char == "(" || char == ")":
			if currentNumber != "" {
				tokens = append(tokens, currentNumber)
				currentNumber = ""
			}
			tokens = append(tokens, char)
		default:
			return nil, fmt.Errorf("invalid character in expression: %s", char)
		}
	}
	
	if currentNumber != "" {
		tokens = append(tokens, currentNumber)
	}
	
	return tokens, nil
}

// shuntingYard implements the shunting yard algorithm to convert infix to postfix notation
func shuntingYard(tokens []string) ([]string, error) {
	var output []string
	var operatorStack []string
	
	for _, token := range tokens {
		switch {
		case isNumber(token):
			output = append(output, token)
		case isOperator(token):
			for len(operatorStack) > 0 && 
				operatorStack[len(operatorStack)-1] != "(" && 
				hasHigherPrecedence(operatorStack[len(operatorStack)-1], token) {
				output = append(output, operatorStack[len(operatorStack)-1])
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
			operatorStack = append(operatorStack, token)
		case token == "(":
			operatorStack = append(operatorStack, token)
		case token == ")":
			for len(operatorStack) > 0 && operatorStack[len(operatorStack)-1] != "(" {
				output = append(output, operatorStack[len(operatorStack)-1])
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
			if len(operatorStack) == 0 || operatorStack[len(operatorStack)-1] != "(" {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			operatorStack = operatorStack[:len(operatorStack)-1] // Remove the "("
		}
	}
	
	for len(operatorStack) > 0 {
		if operatorStack[len(operatorStack)-1] == "(" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, operatorStack[len(operatorStack)-1])
		operatorStack = operatorStack[:len(operatorStack)-1]
	}
	
	return output, nil
}

// isOperator checks if a token is an operator
func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

// isNumber checks if a token is a number
func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// getPrecedence returns the precedence of an operator
func getPrecedence(operator string) int {
	switch operator {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

// hasHigherPrecedence checks if op1 has higher or equal precedence than op2
func hasHigherPrecedence(op1, op2 string) bool {
	return getPrecedence(op1) >= getPrecedence(op2)
}

// ProcessOperation performs the arithmetic operation
func ProcessOperation(operation Operation, arg1, arg2 float64, delay int) (float64, error) {
	// Simulate long computation
	time.Sleep(time.Duration(delay) * time.Millisecond)
	
	switch operation {
	case Addition:
		return arg1 + arg2, nil
	case Subtraction:
		return arg1 - arg2, nil
	case Multiplication:
		return arg1 * arg2, nil
	case Division:
		if arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return arg1 / arg2, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", operation)
	}
}