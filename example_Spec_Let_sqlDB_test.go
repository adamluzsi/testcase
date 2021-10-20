package testcase_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/adamluzsi/testcase"
)

type SupplierWithDBDependency struct {
	DB interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}
}

func (s SupplierWithDBDependency) DoSomething(ctx context.Context) error {
	rows, err := s.DB.QueryContext(ctx, `SELECT 1 = 1`)
	if err != nil {
		return err
	}
	return rows.Close()
}

func ExampleSpec_Let_sqlDB() {
	var t *testing.T
	s := testcase.NewSpec(t)

	var (
		tx = s.Let(`tx`, func(t *testcase.T) interface{} {
			// it is advised to use a persistent db connection between multiple specification runs,
			// because otherwise `go testCase -count $times` can receive random connection failures.
			tx, err := getDBConnection(t).Begin()
			if err != nil {
				t.Fatal(err.Error())
			}
			// testcase.T#Defer will execute the received function after the current testCase edge case
			// where the `tx` testCase variable were accessed.
			t.Defer(tx.Rollback)
			return tx
		})
		supplier = s.Let(`supplier`, func(t *testcase.T) interface{} {
			return SupplierWithDBDependency{DB: tx.Get(t).(*sql.Tx)}
		})
	)

	s.Describe(`#DoSomething`, func(s *testcase.Spec) {
		var (
			ctx = s.Let(`spec`, func(t *testcase.T) interface{} {
				return context.Background()
			})
			subject = func(t *testcase.T) error {
				return supplier.Get(t).(SupplierWithDBDependency).DoSomething(ctx.Get(t).(context.Context))
			}
		)

		s.When(`...`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				//...
			})

			s.Then(`...`, func(t *testcase.T) {
				t.Must.Nil(subject(t))
			})
		})
	})
}

func getDBConnection(t testing.TB) *sql.DB {
	// logic to retrieve cached db connection in the testing environment
	return nil
}
