package stream

import (
	"context"
	"fmt"
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/kind"
	"github.com/vela-security/vela-public/lua"
	audit "github.com/vela-security/vela-audit"
	"net"
	"reflect"
)

var (
	streamTypeOf = reflect.TypeOf((*stream)(nil)).String()
)

type stream struct {
	lua.ProcEx

	cfg *config
	cur config //保存当前启动 为了下次快速启动

	ln *kind.Listener
}

func newStream(cfg *config) *stream {
	obj := &stream{cfg: cfg}
	obj.V(lua.PTInit, streamTypeOf)
	return obj
}

func (st *stream) socket(conn net.Conn) (assert.HTTPStream, error) {
	host := st.cur.remote.Hostname()
	port := st.cur.remote.Port()
	if port == 0 {
		_, port = auxlib.ParseAddr(conn.LocalAddr())
	}

	if port == 0 {
		return nil, fmt.Errorf("invalid stream port")
	}

	return xEnv.Stream("tunnel", map[string]interface{}{
		"network": st.cur.remote.Scheme(),
		"address": fmt.Sprintf("%s:%d", host, port),
	})
}

func (st *stream) pipe(ev *audit.Event) {
	n := st.cur.pipe.Len()
	if n == 0 {
		ev.Put()
		return
	}

	st.cur.pipe.Do(ev, st.cur.co, func(e error) {
		xEnv.Errorf("%s stream pipe fail %v", st.Name(), e)
	})
}

func (st *stream) Code() string {
	return st.cfg.co.CodeVM()
}

func (st *stream) accept(ctx context.Context, conn net.Conn, stop context.CancelFunc) error {
	//toT nt
	bind := st.cur.bind.String()
	remote := st.cur.remote.String()

	ev := audit.NewEvent("chameleon").Alert().High().
		Subject("流式高交互代理蜜罐名命中").
		From(st.Code()).
		Remote(conn.RemoteAddr()).
		Msg("connect %s %s => %s succeed", st.Name(), bind, remote)
	st.pipe(ev)

	var toTn int64

	//接收的数据
	var rev int64

	//报错
	var err error

	//数据通道
	var socket assert.HTTPStream
	socket, err = st.socket(conn)
	ev = audit.NewEvent("chameleon").From(st.Code()).Remote(conn.RemoteAddr()).Alert().High().
		Msg("connect %s %s => %s succeed", st.Name(), bind, remote)

	if err != nil {
		ev.Subject("流式高交互代理蜜罐恶意请求失败").E(err)
		st.pipe(ev)
		return err
	} else {
		ev.Subject("流式高交互代理蜜罐恶意请求成功").E(err)
		st.pipe(ev)
	}

	xEnv.Spawn(0, func() {
		defer func() {
			stop()
			conn.Close()
		}()

		ev = audit.NewEvent("chameleon").From(st.Code()).Remote(conn.RemoteAddr()).Alert().High()

		toTn, err = auxlib.Copy(ctx, socket, conn)
		if err != nil {
			ev.Subject("流式高交互代理蜜罐上游请求关闭").E(err).
				Msg("connect upstream %s %s => %s succeed send:%d", st.Name(), bind, remote, toTn)
			st.pipe(ev)

		} else {
			ev.Subject("流式高交互代理蜜罐上游请求结束").
				Msg("connect upstream %s %s => %s succeed send:%d", st.Name(), bind, remote, toTn)
			st.pipe(ev)
		}
	})

	xEnv.Spawn(0, func() {
		defer func() {
			stop()
			socket.Close()
		}()

		ev = audit.NewEvent("chameleon").From(st.Code()).Remote(conn).Alert().High()
		rev, err = auxlib.Copy(ctx, conn, socket)
		if err != nil {
			ev.Subject("流式高交互代理蜜罐请求关闭").
				Msg("程序名称: %s \n接收远程失败:%d", st.Name(), rev).E(err)
			st.pipe(ev)
		} else {
			ev.Subject("流式高交互代理蜜罐请求结束").
				Msg("程序名称: %s 接收远程结束 数量:%d", st.Name(), rev)
			st.pipe(ev)
		}
	})

	return err
}

func (st *stream) equal() bool {
	if st.cfg.remote.String() != st.cur.remote.String() {
		return false
	}

	if st.cfg.bind.String() != st.cur.bind.String() {
		return false
	}

	return true

}

func (st *stream) Listen() error {

	if st.ln == nil {
		goto conn
	}

conn:
	ln, err := kind.Listen(xEnv, st.cfg.bind)
	if err != nil {
		return err
	}
	st.ln = ln
	return nil
}

func (st *stream) start() (err error) {
	st.cur = *st.cfg
	xEnv.Spawn(100, func() {
		err = st.ln.OnAccept(st.accept)
	})
	return
}

func (st *stream) Start() error {

	if e := st.Listen(); e != nil {
		return e
	}

	return st.start()
}

//func (st *stream) Reload() (err error) {
//	if e := st.Listen(); e != nil {
//		return e
//	}
//
//	st.cur = *st.cfg
//	return nil
//}

func (st *stream) Close() error {
	return st.ln.Close()
}

func (st *stream) Name() string {
	return st.cur.name
}

func (st *stream) Type() string {
	return streamTypeOf
}
