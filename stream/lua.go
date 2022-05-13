package stream

import (
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-public/pipe"
)

var xEnv assert.Environment

/*
	chameleon.stream{
		name = "ssss",
		bind = "tcp://127.0.0.1:3390",
		remote = "tcp://172.31.61.67:3389",
	}
*/

func (s *stream) pipeL(L *lua.LState) int {
	s.cfg.pipe.CheckMany(L, pipe.Seek(0))
	return 0
}

func (s *stream) startL(L *lua.LState) int {
	xEnv.Start(L, s).From(s.Code()).Do()
	return 0
}

func (s *stream) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "pipe":
		return L.NewFunction(s.pipeL)
	case "start":
		return L.NewFunction(s.startL)

	}
	return lua.LNil
}

func newLuaStreamChameleon(L *lua.LState) int {
	cfg := newConfig(L)
	proc := L.NewProc(cfg.name, streamTypeOf)
	if proc.IsNil() {
		proc.Set(newStream(cfg))
	} else {
		proc.Data.(*stream).cfg = cfg
	}

	L.Push(proc)
	return 1
}

func Inject(env assert.Environment, uv lua.UserKV) {
	xEnv = env
	uv.Set("stream", lua.NewFunction(newLuaStreamChameleon))
}
