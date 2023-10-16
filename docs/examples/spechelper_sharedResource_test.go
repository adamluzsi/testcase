// package spechelper
package examples_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/example/mydomain"
	"go.llib.dev/testcase/internal/example/someextres"
)

var (
	sharedGlobalStorageInstanceInit sync.Once
	sharedGlobalStorageInstance     mydomain.Storage // role interface type
)

func getSharedGlobalStorageInstance(tb testing.TB) mydomain.Storage {
	sharedGlobalStorageInstanceInit.Do(func() {
		storage, err := someextres.NewStorage(os.Getenv(`TEST_DATABASE_URL`))
		assert.Must(tb).Nil(err)
		sharedGlobalStorageInstance = storage
	})
	return sharedGlobalStorageInstance
}

var Context = testcase.Var[context.Context]{
	ID: `context`,
	Init: func(t *testcase.T) context.Context {
		return context.Background()
	},
}

var Storage = testcase.Var[mydomain.Storage]{
	ID: `Storage`,
	Init: func(t *testcase.T) mydomain.Storage {
		s := getSharedGlobalStorageInstance(t)
		tx, err := s.BeginTx(Context.Get(t))
		t.Must.Nil(err)
		Context.Set(t, tx)
		t.Defer(s.RollbackTx, tx) // teardown
		return s
	},
}
