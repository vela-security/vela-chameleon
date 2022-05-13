package proxy

import (
	"errors"
	"fmt"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-public/pipe"
)

type config struct {
	Name   string
	Bind   auxlib.URL
	Remote auxlib.URL

	pipe *pipe.Px
	co   *lua.LState
}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cfg := &config{pipe: pipe.New(pipe.Env(xEnv))}
	tab.Range(func(k string, v lua.LValue) {
		switch k {
		case "name":
			cfg.Name = auxlib.CheckProcName(v, L)
		case "bind":
			cfg.Bind = auxlib.CheckURL(v, L)
		case "remote":
			cfg.Remote = auxlib.CheckURL(v, L)
		}
	})

	if e := cfg.verify(); e != nil {
		L.RaiseError("%v", e)
		return nil
	}
	cfg.co = xEnv.Clone(L)
	return cfg
}

func (cfg *config) verify() error {
	if e := auxlib.Name(cfg.Name); e != nil {
		return e
	}

	if cfg.Bind.IsNil() {
		return fmt.Errorf("not found bind url")
	}

	if cfg.Remote.IsNil() {
		return fmt.Errorf("not found remote url")
	}

	switch cfg.Bind.Scheme() {
	case "tcp", "udp":
		//todo

	default:
		return errors.New("invalid bind protocol")
	}

	switch cfg.Remote.Scheme() {
	case "tcp", "udp":
		//todo

	default:
		return errors.New("invalid bind protocol")
	}

	return nil
}
