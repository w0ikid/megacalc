package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceSubmitExpression(t *testing.T) {
	// Create a service with minimal operation times for testing
	svc := NewService(OperationTimes{
		Addition:       1,
		Subtraction:    1,
		Multiplication: 1,
		Division:       1,
	})

	// Test simple expression
	id, err := svc.SubmitExpression("2+2")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Verify expression was created
	expr, found := svc.GetExpression(id)
	assert.True(t, found)
	assert.Equal(t, "2+2", expr.Expression)
	assert.Equal(t, InProcess, expr.Status)

	// Test complex expression
	id, err = svc.SubmitExpression("2+2*2")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Test invalid expression
	_, err = svc.SubmitExpression("2+*2")
	assert.Error(t, err)
}

func TestServiceGetExpressions(t *testing.T) {
	// Create a service with minimal operation times for testing
	svc := NewService(OperationTimes{
		Addition:       1,
		Subtraction:    1,
		Multiplication: 1,
		Division:       1,
	})

	// Add some expressions
	id1, _ := svc.SubmitExpression("2+2")
	id2, _ := svc.SubmitExpression("3*4")

	// Get all expressions
	exprs := svc.GetExpressions()
	assert.Len(t, exprs, 2)

	// Verify expressions are correct
	found1 := false
	found2 := false
	for _, expr := range exprs {
		if expr.ID == id1 {
			found1 = true
			assert.Equal(t, "2+2", expr.Expression)
		}
		if expr.ID == id2 {
			found2 = true
			assert.Equal(t, "3*4", expr.Expression)
		}
	}
	assert.True(t, found1)
	assert.True(t, found2)
}

func TestServiceGetExpression(t *testing.T) {
	// Create a service with minimal operation times for testing
	svc := NewService(OperationTimes{
		Addition:       1,
		Subtraction:    1,
		Multiplication: 1,
		Division:       1,
	})

	// Add an expression
	id, _ := svc.SubmitExpression("2+2")

	// Get the expression
	expr, found := svc.GetExpression(id)
	assert.True(t, found)
	assert.Equal(t, "2+2", expr.Expression)

	// Get a non-existent expression
	_, found = svc.GetExpression("non-existent")
	assert.False(t, found)
}

func TestServiceGetTask(t *testing.T) {
	// Create a service with minimal operation times for testing
	svc := NewService(OperationTimes{
		Addition:       1,
		Subtraction:    1,
		Multiplication: 1,
		Division:       1,
	})

	// Add an expression that creates a simple task
	svc.SubmitExpression("2+2")

	// Get a task
	task, found := svc.GetTask()
	assert.True(t, found)
	assert.NotNil(t, task)
	assert.Equal(t, "2", task.Arg1)
	assert.Equal(t, "2", task.Arg2)
	assert.Equal(t, Addition, task.Operation)
}

func TestServiceSetTaskResult(t *testing.T) {
	// Create a service with minimal operation times for testing
	svc := NewService(OperationTimes{
		Addition:       1,
		Subtraction:    1,
		Multiplication: 1,
		Division:       1,
	})

	// Add an expression
	exprID, _ := svc.SubmitExpression("2+2")

	// Get the task
	task, found := svc.GetTask()
	assert.True(t, found)

	// Set the result
	err := svc.SetTaskResult(task.ID, 4)
	assert.NoError(t, err)

	// Verify the expression is updated
	expr, found := svc.GetExpression(exprID)
	assert.True(t, found)
	assert.Equal(t, Completed, expr.Status)
	assert.NotNil(t, expr.Result)
	assert.Equal(t, 4.0, *expr.Result)
}

func TestProcessOperation(t *testing.T) {
	// Test addition
	result, err := ProcessOperation(Addition, 2, 3, 1)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)

	// Test subtraction
	result, err = ProcessOperation(Subtraction, 5, 3, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)

	// Test multiplication
	result, err = ProcessOperation(Multiplication, 2, 3, 1)
	assert.NoError(t, err)
	assert.Equal(t, 6.0, result)

	// Test division
	result, err = ProcessOperation(Division, 6, 3, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, result)

	// Test division by zero
	_, err = ProcessOperation(Division, 6, 0, 1)
	assert.Error(t, err)

	// Test unknown operation
	_, err = ProcessOperation("unknown", 2, 3, 1)
	assert.Error(t, err)
}