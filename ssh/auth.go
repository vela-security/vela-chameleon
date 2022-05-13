package ssh

import "sync"

type auth struct {
	sync.RWMutex
	data map[string]string
}

func (a *auth) Set(name, pass string) {
	a.Lock()
	a.data[name] = pass
	a.Unlock()
}

func (a *auth) Get(name string) (string, bool) {
	a.RLock()
	defer a.RUnlock()
	v, ok := a.data[name]
	return v, ok
}
