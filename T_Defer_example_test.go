package testcase_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Defer(t *testing.T) {
	s := testcase.NewSpec(t)

	const varName = `db for example is something that needs to defer an action after the test run`
	s.Let(varName, func(t *testcase.T) interface{} {
		db, err := sql.Open(`driverName`, `dataSourceName`)
		require.Nil(t, err)
		t.Defer(db.Close)
		return db
	})

	s.Test(``, func(t *testcase.T) {
		db := t.I(varName).(*sql.DB)
		require.Nil(t, db.Ping())
	})
}
