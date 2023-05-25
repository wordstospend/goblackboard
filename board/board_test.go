package board_test

import (
	"bytes"
	"testing"

	"github.com/wordstospend/goblackboard/board"
)

// notation
// '->' depends on
// ':' generates

// A ->[startkey] : answer which should have as its value the type of startkey

func TestWaitTimeout(t *testing.T) {
	answer := board.Key{
		WatchKey: "answer",
		TypeKey:  "string",
	}

	startkey := board.Key{
		WatchKey: "start",
		TypeKey:  "string",
	}

	A := board.Task{
		Consumer: func(resultChannel board.ResultChannel, argv ...board.Result) {
			resultChannel <- board.Result{Key: answer, Value: []byte(argv[0].Key.TypeKey)}

		},
		TriggerKeys: []board.Key{startkey},
	}

	blackBoard := board.NewBlackBoard()
	blackBoard.AddTask(A)

	results, err := blackBoard.Wait([]board.Key{answer}, 10)
	if err == nil {
		t.Fatal("Timeout did not fire")
	}
	if results != nil {
		t.Fatalf("unexpected results %v", results)
	}

}

func TestDependncyType(t *testing.T) {
	answer := board.Key{
		WatchKey: "answer",
		TypeKey:  "string",
	}

	falseStartKey := board.Key{
		WatchKey: "start",
		TypeKey:  "int",
	}

	startkey := board.Key{
		WatchKey: "start",
		TypeKey:  "string",
	}

	// Task A post testValue after it is triggered by value s
	A := board.Task{
		Consumer: func(resultChannel board.ResultChannel, argv ...board.Result) {
			resultChannel <- board.Result{Key: answer, Value: []byte(argv[0].Key.TypeKey)}

		},
		TriggerKeys: []board.Key{startkey},
	}

	blackBoard := board.NewBlackBoard()
	blackBoard.AddTask(A)

	blackBoard.Publish(falseStartKey, []byte("unused"))
	blackBoard.Publish(startkey, []byte("unused"))

	results, err := blackBoard.Wait([]board.Key{answer}, 1000)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !bytes.Equal(results[0].Value, []byte(startkey.TypeKey)) {
		t.Errorf("a value returned as %s", results)
	}
}

func TestDependency(t *testing.T) {
	a := board.Key{
		WatchKey: "a",
		TypeKey:  "string",
	}

	s := board.Key{
		WatchKey: "s",
		TypeKey:  "string",
	}

	testValue := []byte("ABC")

	// Task A post testValue after it is triggered by value s
	A := board.Task{
		Consumer: func(resultChannel board.ResultChannel, argv ...board.Result) {
			resultChannel <- board.Result{Key: a, Value: testValue}

		},
		TriggerKeys: []board.Key{s},
	}

	blackBoard := board.NewBlackBoard()
	blackBoard.AddTask(A)

	blackBoard.Publish(s, []byte("unused"))

	results, err := blackBoard.Wait([]board.Key{a}, 1000)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !bytes.Equal(results[0].Value, testValue) {
		t.Errorf("a value returned as %s", results)
	}
}

func TestMultiDependency(t *testing.T) {

}
