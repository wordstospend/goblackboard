package board

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Action func(ResultChannel, ...Result)
type Value []byte

type Result struct {
	Key   Key
	Value Value
}

type Key struct {
	WatchKey string
	TypeKey  string
}

func (key Key) Equals(key2 Key) bool {
	return key.WatchKey == key2.WatchKey && key.TypeKey == key2.TypeKey
}

type Task struct {
	Consumer    Action
	TriggerKeys []Key
}

func (task Task) TriggeredBy(key Key) bool {
	for _, tkey := range task.TriggerKeys {
		if tkey.Equals(key) {
			return true
		}
	}
	return false
}

type Slate interface {
	AddTask(Task)                      // Adds a task to the board
	Wait([]Key, int) ([]Result, error) // blocks and waits for the set of keys returns the result
	Publish(Key, Value)                // adds results to the blackboard
}

func NewBlackBoard() Slate {
	board := BlackBoard{
		values:        make(map[string]Value),
		resultChannel: make(chan Result),
		taskChannel:   make(chan Task),      // owned and closed by blackboard
		context:       context.Background(), // tasks are meant to respect the context, it will signal when results channels are closed
		watchers:      []Task{},
		log:           *log.Default(),
	}

	go board.excutionLoop()
	return &board
}

type ResultChannel chan Result
type TaskChannel chan Task

type BlackBoard struct {
	values        map[string]Value
	resultChannel ResultChannel
	taskChannel   TaskChannel
	context       context.Context
	watchers      []Task
	log           log.Logger
}

// A method always to be called within the internal go routine
func (board *BlackBoard) excutionLoop() {
	for {
		select {
		case result := <-board.resultChannel:
			board.publishResult(result)
		case <-board.context.Done():
			return
		case task := <-board.taskChannel:

			board.addTask(task)
		}
	}
}

// adds a task to the board, must be called within the interal go routine
func (board *BlackBoard) addTask(task Task) {
	board.watchers = append(board.watchers, task)
	results := []Result{}
	for _, key := range task.TriggerKeys {
		value, ok := board.values[board.generateKey(key)]
		if ok {
			results = append(results, Result{
				Value: value,
				Key:   key,
			})
		} else {
			return
		}
	}
	go task.Consumer(board.resultChannel, results...)

}

// adds the task to the board
// The task will be started after its dependecies are
// available or immediatly if the dependecies are already
// available
func (board *BlackBoard) AddTask(task Task) {
	board.taskChannel <- task
}

// publish a Result to the blackboard
func (board *BlackBoard) Publish(key Key, value Value) {
	board.resultChannel <- Result{
		Key:   key,
		Value: value,
	}
}

// A blocking function that returns after the Key is published to the blackboard
// or the timeout (ms) expires
func (board *BlackBoard) Wait(keys []Key, timeout int) ([]Result, error) {

	// results of this wait task are passed back to this goroutine via this
	// channel
	waitChan := make(ResultChannel)

	// the task to collect the result of the keys
	task := Task{
		TriggerKeys: keys,
		Consumer: func(resultChannel ResultChannel, results ...Result) {
			for _, result := range results {
				waitChan <- result
			}
			close(waitChan)
		},
	}

	board.AddTask(task)
	// the buffer of resulting values
	resultBuffer := []Result{}
ResultRange:
	for {
		select {
		case <-time.After(time.Millisecond * time.Duration(timeout)):
			return nil, fmt.Errorf("wait timeout with '%v' results board in state %v, watching tasks %v", resultBuffer, board.values, board.watchers)

		case result, ok := <-waitChan:
			if !ok {
				break ResultRange
			}
			resultBuffer = append(resultBuffer, result)
		}

	}

	return resultBuffer, nil
}

func (board *BlackBoard) generateKey(key Key) string {
	return fmt.Sprintf("%s#%s", key.WatchKey, key.TypeKey)
}

type ExecutionContext struct {
	Task    Task
	Results []Result
}

// Add the result to the board and kick off any task that are triggered
// by that added dependecy
func (board *BlackBoard) publishResult(result Result) {
	resultKey := board.generateKey(result.Key)
	board.values[resultKey] = result.Value

	executions := []ExecutionContext{}
WatcherLoop:
	for _, task := range board.watchers {
		// check if we can execute the task
		if task.TriggeredBy(result.Key) {
			results := []Result{}
			for _, key := range task.TriggerKeys {
				value, ok := board.values[board.generateKey(key)]
				if !ok {
					continue WatcherLoop
				}
				results = append(results, Result{Value: value, Key: key})
			}
			executions = append(executions, ExecutionContext{Task: task, Results: results})
		}
	}
	for _, execution := range executions {
		task := execution.Task
		results := execution.Results
		go func() {
			task.Consumer(board.resultChannel, results...)
		}()

	}
}
