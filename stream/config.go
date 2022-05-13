package stream

import (
	"fmt"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-public/pipe"
)

type config struct {
	name   string
	bind   auxlib.URL
	remote auxlib.URL

	pipe *pipe.Px
	co   *lua.LState
}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cfg := &config{
		pipe: pipe.New(),
	}

	tab.Range(func(key string, lv lua.LValue) {
		switch key {
		case "name":
			cfg.name = auxlib.CheckProcName(lv, L)
		case "bind":
			cfg.bind = auxlib.CheckURL(lv, L)
		case "remote":
			cfg.remote = auxlib.CheckURL(lv, L)

		default:
			//todo
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
	if err := auxlib.Name(cfg.name); err != nil {
		return err
	}

	switch cfg.bind.Scheme() {
	case "tcp", "udp", "unix":
		return nil
	default:
		return fmt.Errorf("%s invalid bind url", cfg.name)
	}

	switch cfg.remote.Scheme() {
	case "tcp", "udp", "unix":
		return nil
	default:
		return fmt.Errorf("%s invalid remote url", cfg.name)
	}

	return nil
}
