//package spechelper
package examples_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal/example/mydomain"
	"github.com/adamluzsi/testcase/internal/example/storages"
	"github.com/stretchr/testify/require"
)

var (
	sharedGlobalStorageInstanceInit sync.Once
	sharedGlobalStorageInstance     mydomain.Storage // role interface type
)

func getSharedGlobalStorageInstance(tb testing.TB) mydomain.Storage {
	sharedGlobalStorageInstanceInit.Do(func() {
		storage, err := storages.New(os.Getenv(`TEST_DATABASE_URL`))
		require.Nil(tb, err)
		sharedGlobalStorageInstance = storage
	})
	return sharedGlobalStorageInstance
}

var Context = testcase.Var{
	Name: `context`,
	Init: func(t *testcase.T) interface{} {
		return context.Background()
	},
}

func ContextGet(t *testcase.T) context.Context {
	return Context.Get(t).(context.Context)
}

var Storage = testcase.Var{
	Name: `Storage`,
	Init: func(t *testcase.T) interface{} {
		s := getSharedGlobalStorageInstance(t)
		tx, err := s.BeginTx(ContextGet(t))
		require.Nil(t, err)
		Context.Set(t, tx)
		t.Defer(s.RollbackTx, tx) // teardown
		return s
	},
}

func StorageGet(t *testcase.T) mydomain.Storage {
	return Storage.Get(t).(mydomain.Storage)
}
