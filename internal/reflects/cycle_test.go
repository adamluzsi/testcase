package reflects_test

import (
	"fmt"
	"reflect"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/assertlite"
	"go.llib.dev/testcase/internal/reflects"
)

// TestCycleGuard verifies the public behaviour of reflects.CycleGuard.
func TestCycleGuard(t *testing.T) {
	t.Run("detects pointer cycle", func(t *testing.T) {
		type Node struct{ Next *Node }

		// Build a 2-node cycle.
		n1, n2 := &Node{}, &Node{}
		n1.Next, n2.Next = n2, n1

		var g reflects.CycleGuard

		// First visit -> no cycle.
		v := reflect.ValueOf(n1)
		assertlite.True(t, g.CheckAndMark(v), "first visit must not flag a cycle")

		// Second visit to the *same* address -> cycle.
		assertlite.False(t, g.CheckAndMark(v), "second visit should flag a cycle")
	})

	t.Run("interface is ignored as an interface value is what can have cycle, not the interface type itself", func(t *testing.T) {
		var m string = "foo"
		var i any = m
		var g reflects.CycleGuard

		v := reflect.ValueOf(&i).Elem()

		assertlite.True(t, g.CheckAndMark(v))
		assertlite.False(t, g.CheckAndMark(v), "interface is not something that is considered as seen",
			"this is very important because when visiting an interface, and marking the top level as seen, we should not get seen in the internal value itself")

		assertlite.True(t, g.CheckAndMark(v.Elem()))
	})

	t.Run("detects self-referential map", func(t *testing.T) {
		m := make(map[string]any)
		m["self"] = m

		var g reflects.CycleGuard
		v := reflect.ValueOf(m)

		assertlite.True(t, g.CheckAndMark(v))
		assertlite.False(t, g.CheckAndMark(v))
		mapSelfVal := v.MapIndex(reflect.ValueOf("self"))
		assertlite.Equal(t, mapSelfVal.Kind(), reflect.Interface)
		assertlite.False(t, mapSelfVal.IsNil())
		fmt.Println("---")
		assertlite.False(t, g.CheckAndMark(mapSelfVal), "self reference within the same map must trigger cycle")
	})

	t.Run("slice containing itself", func(t *testing.T) {
		s := make([]any, 1)
		s[0] = s

		var g reflects.CycleGuard
		v := reflect.ValueOf(s)

		assertlite.True(t, g.CheckAndMark(v))
		assertlite.False(t, g.CheckAndMark(v), "identical slice triggers cycle")
		elem := v.Index(0)
		assertlite.Equal(t, elem.Kind(), reflect.Interface)
		assertlite.False(t, elem.IsNil())
		assertlite.False(t, g.CheckAndMark(elem), "recursive value should be detected as cycle")
	})

	t.Run("independent guards are isolated", func(t *testing.T) {
		type N struct{ Next *N }
		root := &N{}
		root.Next = root // simple cycle

		v := reflect.ValueOf(root)

		var (
			guardA reflects.CycleGuard
			guardB reflects.CycleGuard
		)

		assert.True(t, guardA.CheckAndMark(v), "A first visit ok")
		assertlite.False(t, guardA.CheckAndMark(v), "A detects cycle on repeat")
		assert.True(t, guardB.CheckAndMark(v), "B is independent of A")
	})

	// Replace the failing test with these corrected versions

	t.Run("channels can be reliably cycle-detected", func(t *testing.T) {
		var g reflects.CycleGuard

		ch := make(chan int)
		chVal := reflect.ValueOf(ch)

		// The cycle guard should handle this gracefully by not detecting cycles
		// since it can't get a stable address
		result1 := g.CheckAndMark(chVal)
		result2 := g.CheckAndMark(chVal)

		// Both calls should return false since no stable address is available
		assert.True(t, result1, "first channel visit should not detect cycle")
		assertlite.False(t, result2, "second channel visit should not detect cycle due to unstable address")
	})

	t.Run("functions can be reliably cycle-detected", func(t *testing.T) {
		var g reflects.CycleGuard

		fn := func() {}
		fnVal := reflect.ValueOf(fn)

		// Similar to channels, functions don't provide stable addresses
		result1 := g.CheckAndMark(fnVal)
		result2 := g.CheckAndMark(fnVal)

		assert.True(t, result1, "first function visit should not detect cycle")
		assertlite.False(t, result2, "second function visit should be detect cycle")
	})

	t.Run("types with stable addresses work correctly", func(t *testing.T) {
		var g reflects.CycleGuard

		// Test with types that DO have stable addresses
		m := make(map[string]int)
		m["key"] = 42

		s := make([]int, 1)
		s[0] = 42

		mapVal := reflect.ValueOf(m)
		sliceVal := reflect.ValueOf(s)

		// These should work because maps and slices have stable UnsafePointer addresses
		assertlite.True(t, g.CheckAndMark(mapVal), "first map visit should not detect cycle")
		assertlite.False(t, g.CheckAndMark(mapVal), "second map visit should detect cycle")

		assertlite.True(t, g.CheckAndMark(sliceVal), "first slice visit should not detect cycle")
		assertlite.False(t, g.CheckAndMark(sliceVal), "second slice visit should detect cycle")
	})

	t.Run("primitive types are excluded from cycle guarding", func(t *testing.T) {
		// These types are not addressable and not supported by UnsafePointer
		primitives := []any{
			bool(true),
			int(42),
			int8(1),
			int16(2),
			int32(3),
			int64(4),
			uint(5),
			uint8(6),
			uint16(7),
			uint32(8),
			uint64(9),
			uintptr(10),
			float32(1.1),
			float64(2.2),
			complex64(1 + 2i),
			complex128(3 + 4i),
		}

		for _, val := range primitives {
			val := val
			name := reflect.TypeOf(val).String()

			t.Run(name, func(t *testing.T) {
				var g reflects.CycleGuard

				v := reflect.ValueOf(val)
				result1 := g.CheckAndMark(v)
				result2 := g.CheckAndMark(v)

				assert.True(t, result1, assert.MessageF("first visit to primitive %s should not detect cycle", name))
				assert.True(t, result2, assert.MessageF("second visit to primitive %s should not detect cycle (no stable address)", name))
			})
		}
	})

	t.Run("pointer itself is not considered equal with the value itself it points to", func(t *testing.T) {
		type T struct{ V int }
		var (
			n T             = T{V: 42}
			p *T            = &n
			r reflect.Value = reflect.ValueOf(p)
			g reflects.CycleGuard
		)

		assertlite.Equal(t, r.Kind(), reflect.Pointer)
		assertlite.True(t, g.CheckAndMark(r))
		assertlite.False(t, g.CheckAndMark(r))
		assertlite.Equal(t, r.Elem().Kind(), reflect.Struct)
		assertlite.True(t, g.CheckAndMark(r.Elem()))
		assertlite.False(t, g.CheckAndMark(r.Elem()))
	})

	t.Run("reflect.Type visit", func(t *testing.T) {
		var (
			t1 = reflect.TypeOf((*int)(nil)).Elem()
			t2 = reflect.TypeOf((*string)(nil)).Elem()

			v1 = reflect.ValueOf(t1)
			v2 = reflect.ValueOf(t2)
		)

		var g reflects.CycleGuard

		assertlite.True(t, g.CheckAndMark(v1))
		assertlite.False(t, g.CheckAndMark(v1))

		assertlite.True(t, g.CheckAndMark(v2))
		assertlite.False(t, g.CheckAndMark(v2))
	})

	t.Run("nested structs", func(t *testing.T) {
		var g reflects.CycleGuard

		type X struct{}

		type T struct {
			X1 X
			X2 X
		}

		var v T

		rv := reflect.ValueOf(&v).Elem()
		assertlite.True(t, g.CheckAndMark(rv))
		assertlite.False(t, g.CheckAndMark(rv))

		x1 := rv.FieldByName("X1")
		assertlite.Equal(t, x1.Kind(), reflect.Struct)
		assertlite.True(t, g.CheckAndMark(x1))
		assertlite.False(t, g.CheckAndMark(x1))

		x2 := rv.FieldByName("X2")
		assertlite.Equal(t, x2.Kind(), reflect.Struct)
		assertlite.True(t, g.CheckAndMark(x2))
		assertlite.False(t, g.CheckAndMark(x2))

		t.Log("T  ", uintptr(rv.Addr().UnsafePointer()))
		t.Log("T.X1", uintptr(x1.Addr().UnsafePointer()))
		t.Log("T.X2", uintptr(x2.Addr().UnsafePointer()))
	})
}
