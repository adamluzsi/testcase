package internal

type Stack[T any] []T

// IsEmpty check if stack is empty
func (s *Stack[T]) IsEmpty() bool {
	return len(*s) == 0
}

// Push a new value onto the stack
func (s *Stack[T]) Push(v T) {
	*s = append(*s, v) // Simply append the new value to the end of the stack
}

// Pop remove and return top element of stack. Return false if stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		return *new(T), false
	}
	index := len(*s) - 1   // Get the index of the top most element.
	element := (*s)[index] // Index into the slice and obtain the element.
	*s = (*s)[:index]      // Remove it from the stack by slicing it off.
	return element, true
}

// Last returns the last stack element
func (s *Stack[T]) Last() (T, bool) {
	if s.IsEmpty() {
		return *new(T), false
	}
	return (*s)[(len(*s) - 1)], true
}
