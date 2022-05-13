package ssh

import (
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/lua"
)

type config struct {
	name    string
	bind    string
	prompt  string
	version string
	code    string
}

func def(L *lua.LState) *config {
	return &config{
		bind:    "0.0.0.0:2222",
		prompt:  "root#",
		version: "OpenSSH_7.4",
		code:    L.CodeVM(),
	}
}

func newConfig(L *lua.LState) *config {
	tab := L.CheckTable(1)
	cfg := def(L)

	tab.ForEach(func(key lua.LValue, val lua.LValue) {
		switch key.String() {
		case "name":
			cfg.name = val.String()
		case "bind":
			cfg.bind = val.String()
		case "prompt":
			cfg.prompt = val.String()
		case "version":
			cfg.version = val.String()

		default:
			L.RaiseError("not found %s key", key.String())
		}
	})

	if e := cfg.verify(); e != nil {
		L.RaiseError("%v", e)
		return nil
	}

	return cfg

}

func (cfg *config) verify() error {
	if e := auxlib.Name(cfg.name); e != nil {
		return e
	}

	return nil
}

func (cfg *config) toSSH(h Handler, p PasswordHandler) *Server {
	return &Server{Addr: cfg.bind, Handler: h, PasswordHandler: p}
}
