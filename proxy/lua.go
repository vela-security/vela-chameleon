package proxy

import (
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-public/pipe"
)

var xEnv assert.Environment

func (p *proxyGo) pipeL(L *lua.LState) int {
	p.cfg.pipe.CheckMany(L, pipe.Seek(0))
	return 0
}

func (p *proxyGo) startL(L *lua.LState) int {
	xEnv.Start(L, p).From(p.CodeVM()).Do()
	return 0
}

func (p *proxyGo) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "pipe":
		return L.NewFunction(p.pipeL)
	case "start":
		return L.NewFunction(p.startL)

	}

	return lua.LNil
}

/*
	chameleon.proxy{
		name = "xxxxx",
		bind = "tcp://127.0.0.1:3309",
		remote = "tcp://172.31.61.67:3389"
	}

*/
func newLuaProxyChameleon(L *lua.LState) int {
	cfg := newConfig(L)

	proc := L.NewProc(cfg.Name, proxyTypeOf)
	if proc.IsNil() {
		proc.Set(newProxyGo(cfg))
	} else {
		proc.Data.(*proxyGo).cfg = cfg
	}

	L.Push(proc)
	return 1
}

func Inject(env assert.Environment, uv lua.UserKV) {
	xEnv = env
	uv.Set("proxy", lua.NewFunction(newLuaProxyChameleon))
}
