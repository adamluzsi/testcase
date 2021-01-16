package testcase_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Defer() {
	var t *testing.T
	s := testcase.NewSpec(t)

	const varName = `db for example is something that needs to defer an action after the testCase run`
	db := s.Let(varName, func(t *testcase.T) interface{} {
		db, err := sql.Open(`driverName`, `dataSourceName`)

		// asserting error here with the *testcase.T ensure that the testCase will don't have some spooky failure.
		require.Nil(t, err)

		// db.Close() will be called after the current testCase case reach the teardown hooks
		t.Defer(db.Close)

		// check if connection is OK
		require.Nil(t, db.Ping())

		// return the verified db instance for the caller
		// this db instance will be memorized during the runtime of the testCase case
		return db
	})

	s.Test(`a simple testCase case`, func(t *testcase.T) {
		db := db.Get(t).(*sql.DB)
		require.Nil(t, db.Ping()) // just to do something with it.
	})
}
