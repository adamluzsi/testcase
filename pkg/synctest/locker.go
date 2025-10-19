package synctest

import "sync"

type nopLocker struct{}

func (*nopLocker) Lock() {}

func (*nopLocker) Unlock() {}

type multiLocker []sync.Locker

func (ls multiLocker) Lock() {
	for _, l := range ls {
		l.Lock()
	}
}

func (ls multiLocker) Unlock() {
	for _, l := range ls {
		l.Unlock()
	}
}
