package testcase_test

import (
	"database/sql"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Defer() {
	var t *testing.T
	s := testcase.NewSpec(t)

	const varName = `db for example is something that needs to defer an action after the testCase run`
	db := testcase.Let(s, varName, func(t *testcase.T) *sql.DB {
		db, err := sql.Open(`driverName`, `dataSourceName`)

		// asserting error here with the *testcase.T ensure that the testCase will don't have some spooky failure.
		t.Must.Nil(err)

		// db.Close() will be called after the current test case reach the teardown hooks
		t.Defer(db.Close)

		// check if connection is OK
		t.Must.Nil(db.Ping())

		// return the verified db instance for the caller
		// this db instance will be memorized during the runtime of the test case
		return db
	})

	s.Test(`a simple test case`, func(t *testcase.T) {
		db := db.Get(t)
		t.Must.Nil(db.Ping()) // just to do something with it.
	})
}
