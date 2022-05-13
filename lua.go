package chameleon

import (
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-chameleon/mysql"
	"github.com/vela-security/vela-chameleon/proxy"
	"github.com/vela-security/vela-chameleon/ssh"
	"github.com/vela-security/vela-chameleon/stream"
)

func WithEnv(env assert.Environment) {
	uv := lua.NewUserKV()
	proxy.Inject(env, uv)
	stream.Inject(env, uv)
	mysql.Inject(env, uv)
	ssh.Inject(env, uv)
	env.Global("chameleon", uv)
}
