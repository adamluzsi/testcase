package testcase

import "sync"

// Global configures all *Spec which is made afterward of this call.
// If you need a Spec#Before that runs in configured in every Spec, use this function.
// It can be called multiple times, and then configurations will stack.
var Global global

type global struct {
	mutex     sync.RWMutex
	beforeFns []tBlock
}

func (gc *global) Before(block tBlock) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()
	gc.beforeFns = append(gc.beforeFns, block)
}

func applyGlobal(s *Spec) {
	Global.mutex.RLock()
	defer Global.mutex.RUnlock()
	for _, block := range Global.beforeFns {
		s.Before(block)
	}
}
