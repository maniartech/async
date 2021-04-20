package async_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maniartech/async"
	"github.com/stretchr/testify/assert"
)

func TestGoPromiseBase(t *testing.T) {
	promise := async.Go(processAsync, "A", 1000)

	isPromise := false

	if _, ok := interface{}(promise).(*async.Promise); ok {
		isPromise = true
	}

	assert.Equal(t, true, isPromise)
	assert.Equal(t, true, promise.NotStarted())
	assert.Equal(t, false, promise.Pending())
	assert.Equal(t, false, promise.Finished())

}

func TestGoPromise(t *testing.T) {
	promise := async.Go(processAsync, "A", 1000)
	result, err := promise.Await()

	assert.Equal(t, true, promise.Finished())

	assert.Equal(t, "A", result)
	assert.Equal(t, nil, err)

	promise = async.Go(processAsync, "A", 1000, errors.New("invalid-action"))
	result, err = promise.Await()

	assert.Equal(t, true, promise.Finished())

	assert.Equal(t, nil, result)
	assert.EqualError(t, err, "invalid-action")

}

func TestBatchGo(t *testing.T) {

	vals := make([]string, 0)
	newCB := func() func(string) {
		return func(s string) {
			vals = append(vals, s)
		}
	}

	async.GoP(
		async.Go(processAsync, "A", 3000, newCB()),
		async.Go(processAsync, "B", 2000, newCB()),
		async.GoQ( // Calls Go routines in queue!
			async.Go(processAsync, "C", 1000, newCB()),
			async.Go(processAsync, "D", 500, newCB()),
			async.Go(processAsync, "E", 100, newCB()),
		),
		async.GoP(
			async.Go(processAsync, "F", 200, newCB()),
			async.Go(processAsync, "G", 0, newCB()),
		),
	).Await()

	assert.Equal(t, "G,F,C,D,E,B,A", strings.Join(vals, ","))
}

func processAsync(p *async.Promise, args ...interface{}) {
	s := args[0].(string)
	ms := args[1].(int)

	time.Sleep(time.Duration(ms) * time.Millisecond)

	// If callback is supplied, call it by passing s!
	if len(args) == 3 {
		switch args[2].(type) {
		case func(string):
			p.Done(s)
			cb := args[2].(func(string))
			cb(s)
		case error:
			p.Done(args[2])
		default:
			p.Done(s)
		}
		return
	}
	p.Done(s)
}
