package testcase

import (
	"os"
	"sync"
	"testing"
)

var chdirconf = chdirConfig{
	register: make(map[string]struct{}),
}

type chdirConfig struct {
	chLock   sync.Mutex
	regLock  sync.RWMutex
	register map[string]struct{}
}

func (conf *chdirConfig) Lock(tb testing.TB) {
	if conf.IsTBRegistered(tb) {
		return
	}

	chdirconf.regLock.Lock()
	defer chdirconf.regLock.Unlock()

	if conf.isTBRegistered(tb) {
		return
	}

	conf.chLock.Lock()
	tb.Cleanup(conf.chLock.Unlock)
	conf.registerTB(tb)

	tb.Cleanup(func() {
		conf.deregisterTB(tb)
	})
}

func (conf *chdirConfig) IsTBRegistered(tb testing.TB) bool {
	chdirconf.regLock.RLock()
	defer chdirconf.regLock.RUnlock()
	return conf.isTBRegistered(tb)
}

func (conf *chdirConfig) deregisterTB(tb testing.TB) {
	chdirconf.regLock.Lock()
	defer chdirconf.regLock.Unlock()
	delete(conf.register, tb.Name())
}

func (conf *chdirConfig) registerTB(tb testing.TB) {
	if conf.register == nil {
		conf.register = make(map[string]struct{})
	}
	conf.register[tb.Name()] = struct{}{}
}

func (conf *chdirConfig) isTBRegistered(tb testing.TB) bool {
	if conf.register == nil {
		return false
	}
	_, isRegistered := conf.register[tb.Name()]
	return isRegistered
}

func Chdir(tb testing.TB, dir string) {
	chdirconf.Lock(tb)

	pwd, err := os.Getwd()
	if err != nil {
		tb.Fatal(err.Error())
	}
	tb.Cleanup(func() {
		if err := os.Chdir(pwd); err != nil {
			tb.Fatal(err.Error())
		}
	})
	if err := os.Chdir(dir); err != nil {
		tb.Fatal(err.Error())
	}
}
