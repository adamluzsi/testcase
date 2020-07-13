package testcase_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleT_HasTag() {
	var t *testing.T
	var s = testcase.NewSpec(t)

	s.Let(`db`, func(t *testcase.T) interface{} {
		db, err := sql.Open(`driverName`, `dataSourceName`)
		require.Nil(t, err)

		if t.HasTag(`black box`) {
			// tests with black box  use http test server or similar things and high level tx management not maintainable.
			t.Defer(db.Close)
			return db
		}

		tx, err := db.BeginTx(context.Background(), nil)
		require.Nil(t, err)
		t.Defer(tx.Rollback)
		return tx
	})
}
