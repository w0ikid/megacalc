package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/w0ikid/megacalc/internal/service"
)

// Agent represents a computational agent
type Agent struct {
	orchestratorURL string
	computingPower  int
	client          *http.Client
}

// TaskResponse represents a task response from the orchestrator
type TaskResponse struct {
	Task *service.Task `json:"task,omitempty"`
}

// TaskResultRequest represents a request to set a task result
type TaskResultRequest struct {
	ID     string  `json:"id" binding:"required"`
	Result float64 `json:"result" binding:"required"`
}

// NewAgent creates a new agent
func NewAgent(orchestratorURL string, computingPower int) *Agent {
	return &Agent{
		orchestratorURL: orchestratorURL,
		computingPower:  computingPower,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Start starts the agent with the specified computing power
func (a *Agent) Start() {
	log.Printf("Starting agent with %d computing power", a.computingPower)
	
	var wg sync.WaitGroup
	
	// Start computing goroutines
	for i := 0; i < a.computingPower; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			a.worker(workerID)
		}(i)
	}
	
	wg.Wait()
}

// worker is the main worker loop
func (a *Agent) worker(id int) {
	log.Printf("Worker %d started", id)
	
	for {
		// Get a task
		task, err := a.getTask()
		if err != nil {
			log.Printf("Worker %d: Error getting task: %v, retrying in 1 second", id, err)
			time.Sleep(1 * time.Second)
			continue
		}
		
		// No task available, wait a bit and try again
		if task == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		
		log.Printf("Worker %d: Processing task %s: %s %s %s", id, task.ID, task.Arg1, task.Operation, task.Arg2)
		
		// Process the task
		result, err := a.processTask(task)
		if err != nil {
			log.Printf("Worker %d: Error processing task %s: %v", id, task.ID, err)
			continue
		}
		
		// Submit the result
		err = a.submitResult(task.ID, result)
		if err != nil {
			log.Printf("Worker %d: Error submitting result for task %s: %v", id, task.ID, err)
			continue
		}
		
		log.Printf("Worker %d: Completed task %s with result %.2f", id, task.ID, result)
	}
}

// getTask gets a task from the orchestrator
func (a *Agent) getTask() (*service.Task, error) {
	url := fmt.Sprintf("%s/internal/task", a.orchestratorURL)
	
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		// No task available
		return nil, nil
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}
	
	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, err
	}
	
	return taskResp.Task, nil
}

// processTask processes a task
func (a *Agent) processTask(task *service.Task) (float64, error) {
	// Parse the arguments
	arg1, err := parseArg(task.Arg1)
	if err != nil {
		return 0, fmt.Errorf("invalid arg1: %v", err)
	}
	
	arg2, err := parseArg(task.Arg2)
	if err != nil {
		return 0, fmt.Errorf("invalid arg2: %v", err)
	}
	
	// Process the operation
	return service.ProcessOperation(task.Operation, arg1, arg2, task.OperationTime)
}

// submitResult submits the result to the orchestrator
func (a *Agent) submitResult(taskID string, result float64) error {
	url := fmt.Sprintf("%s/internal/task", a.orchestratorURL)
	
	resultReq := TaskResultRequest{
		ID:     taskID,
		Result: result,
	}
	
	jsonData, err := json.Marshal(resultReq)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}
	
	return nil
}

// parseArg parses an argument, which could be a task ID or a numerical value
func parseArg(arg string) (float64, error) {
	return strconv.ParseFloat(arg, 64)
}