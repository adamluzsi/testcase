package reflects

import (
	"fmt"
	"reflect"
)

// CycleGuard tracks visited addresses to prevent infinite loops
type CycleGuard struct {
	visited map[visitedCycleGuard]struct{}
}

type visitedCycleGuard struct {
	Addr uintptr
	Type reflect.Type
	Kind reflect.Kind
}

// CheckAndMark returns true if this value has already been visited (cycle detected)
// Returns false if it's safe to proceed (and marks it as visited)
func (g *CycleGuard) CheckAndMark(v reflect.Value) (ok bool) {
	fmt.Println("1")
	if !g.needsCycleCheck(v) {
		fmt.Println("1/A")
		return true // No cycle possible for this type
	}
	fmt.Println("1/B")

	fmt.Println(v.Kind().String())
	fmt.Println(v.CanAddr())
	fmt.Println("2")
	id, ok := g.toID(v)
	if !ok {
		fmt.Println("2/A")
		return true // Couldn't get address, assume safe
	}
	fmt.Println("2/B")

	fmt.Println("???")

	// Check if already visited
	if g.seen(id) {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
			return false // Cycle detected!
		default: // self detected
			return true
		}
	}

	// Mark as visited
	g.mark(id)
	return true
}

func (g *CycleGuard) Unmark(v reflect.Value) {
	if g.visited == nil {
		return
	}
	if !g.needsCycleCheck(v) {
		return // No cycle possible for this type
	}
	id, ok := g.toID(v)
	if !ok {
		return // Couldn't get address, assume safe
	}
	delete(g.visited, id)
}

// needsCycleCheck returns true for types that can form cycles
func (g *CycleGuard) mark(id visitedCycleGuard) {
	if g.visited == nil {
		g.visited = make(map[visitedCycleGuard]struct{})
	}
	g.visited[id] = struct{}{}
}

func (g *CycleGuard) seen(id visitedCycleGuard) bool {
	if g.visited == nil {
		return false
	}
	_, seen := g.visited[id]
	return seen
}

func (g *CycleGuard) needsCycleCheck(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Struct:
		return v.CanAddr() // Only if addressable
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return !v.IsNil()
	default: // primitive value
		return false
	}
}

// toID returns a stable address for cycle detection
func (g *CycleGuard) toID(v reflect.Value) (visitedCycleGuard, bool) {
	addr, ok := LookupUnsafeAddr(v)
	if !ok {
		var zero visitedCycleGuard
		return zero, false
	}
	return visitedCycleGuard{Addr: addr}, true
}

func LookupUnsafeAddr(v reflect.Value) (uintptr, bool) {
	// For addressable values, use UnsafeAddr
	if v.CanAddr() {
		return v.UnsafeAddr(), true
	} else if v.Kind() == reflect.Interface && !v.IsNil() {
		if elem := v.Elem(); elem.CanAddr() {
			return elem.UnsafeAddr(), true
		}
	}
	// For specific non-addressable types, use UnsafePointer
	if canUseUnsafePointer(v) {
		return uintptr(v.UnsafePointer()), true
	}
	return 0, false // No stable address available
}

var canUseUnsafePointerKinds = map[reflect.Kind]struct{}{
	reflect.Chan:          {},
	reflect.Func:          {},
	reflect.Map:           {},
	reflect.Pointer:       {},
	reflect.Slice:         {},
	reflect.String:        {},
	reflect.UnsafePointer: {},
}

// canUseUnsafePointer returns true if v.UnsafePointer() can be called
func canUseUnsafePointer(v reflect.Value) bool {
	_, ok := canUseUnsafePointerKinds[v.Kind()]
	return ok
}
