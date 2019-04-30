package testrun

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)


func NewCTX() *CTX {
	return &CTX{
		steps: make([]func(*testing.T), 0),
		vars:  make(Vars),
		lock:  &sync.Mutex{},
	}
}

// CTX provide you a struct that make building nested test context easy with the core T#Run function.
// ideal for synchronous nested test blocks
type CTX struct {
	steps []func(*testing.T)
	lock  *sync.Mutex
	vars  Vars
}

func (ctx *CTX) Step(step func(t *testing.T)) func() {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	ctx.steps = append(ctx.steps, step)

	return func() {
		ctx.lock.Lock()
		defer ctx.lock.Unlock()

		ctx.steps = ctx.steps[:len(ctx.steps)-1]
	}
}

func (ctx *CTX) Setup(t *testing.T) Vars {
	for _, fn := range ctx.steps {
		fn(t)
	}

	newVars := make(Vars)

	for k, v := range ctx.vars {
		newVars[k] = v
	}

	return newVars
}

// Vars

func (ctx *CTX) Let(varName string, letBlock func(Vars) interface{}) func() {
	return ctx.Step(func(t *testing.T) {
		ctx.vars[varName] = letBlock
	})
}

type Vars map[string]func(Vars) interface{}

func (vars Vars) Get(varName string, dstPtr interface{}) {

	fn, found := vars[varName]

	if !found {
		panic(vars.errorFor(varName))
	}

	src := fn(vars)

	value := reflect.ValueOf(src)

	if value.Kind() != reflect.Ptr {
		ptr := reflect.New(reflect.TypeOf(src))
		ptr.Elem().Set(value)
		value = ptr
	}

	reflect.ValueOf(dstPtr).Elem().Set(value.Elem())

}

func (vars Vars) errorFor(varName string) error {

	var msgs []string
	msgs = append(msgs, fmt.Sprintf(`the following variable not found: %s`, varName))

	var keys []string
	for k, _ := range vars {
		keys = append(keys, k)
	}

	msgs = append(msgs, fmt.Sprintf(`did you mean one from these? %s`, strings.Join(keys, `, `)))

	return errors.New(strings.Join(msgs, ". "))

}
