package testcase_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_sqlDB() {
	var t *testing.T
	s := testcase.NewSpec(t)

	// I highly recommend to use *sql.Tx when it is possible for testing.
	// it allows you to have easy teardown
	s.Let(`tx`, func(t *testcase.T) interface{} {
		// it is advised to use a persistent db connection between multiple specification runs,
		// because otherwise `go test -count $times` can receive random connection failures.
		tx, err := getDBConnection(t).Begin()
		if err != nil {
			t.Fatal(err.Error())
		}
		// testcase.T#Defer will execute the received function after the current test edge case
		// where the `tx` test variable were accessed.
		t.Defer(tx.Rollback)
		return tx
	})

	s.When(`something to be prepared in the db`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			_, err := t.I(`tx`).(*sql.Tx).Exec(`INSERT INTO "table" ("column") VALUES ($1)`, `value`)
			require.Nil(t, err)
		})

		s.Then(`something will happen`, func(t *testcase.T) {
			// ...
		})
	})

}

func getDBConnection(t testing.TB) *sql.DB {
	// logic to retrieve cached db connection in the testing environment
	return nil
}
