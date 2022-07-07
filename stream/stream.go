package stream

import (
	"context"
	"fmt"
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/kind"
	"github.com/vela-security/vela-public/lua"
	risk "github.com/vela-security/vela-risk"
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

func (st *stream) pipe(ev *risk.Event) {
	if st.cur.log {
		ev.Log()
	}

	st.cur.pipe.Do(ev, st.cur.co, func(e error) {
		xEnv.Errorf("%s stream pipe fail %v", st.Name(), e)
	})

	if st.cur.alert && ev.Alert {
		ev.Send()
	}
}

func (st *stream) Code() string {
	return st.cfg.co.CodeVM()
}

func (st *stream) accept(ctx context.Context, conn net.Conn, stop context.CancelFunc) error {
	//toT nt
	bind := st.cur.bind.String()
	remote := st.cur.remote.String()

	ev := risk.HoneyPot()
	ev.From(st.CodeVM())
	ev.Remote(conn.RemoteAddr())
	ev.Local(conn.LocalAddr())
	ev.Subjectf("流式代理蜜罐命中")
	ev.Payload = bind
	st.pipe(ev)

	var toTn int64

	//接收的数据
	var rev int64

	//报错
	var err error

	//数据通道
	var socket assert.HTTPStream
	socket, err = st.socket(conn)

	xEnv.Spawn(0, func() {
		defer func() {
			stop()
			conn.Close()
		}()
		toTn, err = auxlib.Copy(ctx, socket, conn)
		xEnv.Infof("stream %s proxy send %v data:%d", st.Name(), remote, toTn)
	})

	xEnv.Spawn(0, func() {
		defer func() {
			stop()
			socket.Close()
		}()

		rev, err = auxlib.Copy(ctx, conn, socket)
		xEnv.Infof("stream %s proxy recv  %s data:%d", st.Name(), remote, rev)
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
