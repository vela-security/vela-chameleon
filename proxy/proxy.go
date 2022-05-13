package proxy

import (
	"context"
	"fmt"
	"github.com/vela-security/vela-public/auxlib"
	"github.com/vela-security/vela-public/kind"
	"github.com/vela-security/vela-public/lua"
	audit "github.com/vela-security/vela-audit"
	"net"
	"reflect"
	"time"
)

var proxyTypeOf = reflect.TypeOf((*proxyGo)(nil)).String()

type proxyGo struct {
	lua.ProcEx
	cfg *config
	cur config
	ln  *kind.Listener
}

func newProxyGo(cfg *config) *proxyGo {
	p := &proxyGo{cfg: cfg}
	p.V(lua.PTInit, proxyTypeOf)
	return p
}

func (p *proxyGo) Name() string {
	return p.cur.Name
}

func (p *proxyGo) Code() string {
	return p.cfg.co.CodeVM()
}

func (p *proxyGo) equal() bool {
	if p.cur.Bind.String() != p.cfg.Bind.String() {
		return false
	}

	if p.cur.Remote.String() != p.cfg.Remote.String() {
		return false
	}

	return true
}

func (p *proxyGo) Listen() error {

	if p.ln == nil {
		goto conn
	}

conn:
	ln, err := kind.Listen(xEnv, p.cfg.Bind)
	if err != nil {
		return err
	}
	p.ln = ln
	p.cur = *p.cfg

	return nil
}

func (p *proxyGo) Start() error {

	if e := p.Listen(); e != nil {
		return e
	}

	var err error
	xEnv.Spawn(100, func() { err = p.ln.OnAccept(p.accept) })
	return err
}

//func (p *proxyGo) Reload() error {
//	return p.Listen()
//}

func (p *proxyGo) Close() error {
	e := p.ln.Close()
	return e
}

func (p *proxyGo) dail(conn net.Conn) (net.Conn, error) {

	host := p.cur.Remote.Hostname()
	port := p.cur.Remote.Port()

	if port == 0 {
		_, port = auxlib.ParseAddr(conn.LocalAddr())
	}

	if port == 0 {
		return nil, fmt.Errorf("invalid stream port")
	}

	d := net.Dialer{Timeout: 2 * time.Second}
	return d.Dial(p.cur.Remote.Scheme(), fmt.Sprintf("%s:%d", host, port))
}

func (p *proxyGo) pipe(ev *audit.Event) {
	n := p.cur.pipe.Len()
	if n == 0 {
		ev.Log().Put()
		return
	}

	p.cur.pipe.Do(ev, p.cur.co, func(e error) {
		xEnv.Errorf("%s pipe call fail %v", p.Name(), e)
	})
}

func (p *proxyGo) accept(ctx context.Context, conn net.Conn, stop context.CancelFunc) error {

	ev := audit.NewEvent("chameleon").Alert().High().
		Subject("高交互代理蜜罐有新的请求").
		From(p.cur.co.CodeVM()).
		Remote(conn.RemoteAddr())

	dst, err := p.dail(conn)
	if err != nil {
		ev.Msg("%s 服务端口:%s 后端地址:%s 链接失败", p.Name(), conn.LocalAddr().String(),
			p.cfg.Remote).E(err)
		p.pipe(ev)
		return err

	} else {
		ev.Msg("%s 服务端口:%s 后端地址:%s 链接成功", p.Name(), conn.LocalAddr().String(),
			dst.RemoteAddr().String())
		p.pipe(ev)
	}

	xEnv.Spawn(50, func() {
		defer func() {
			stop()
			conn.Close()
		}()

		var toTn int64
		ev = audit.NewEvent("chameleon").From(p.cur.co.CodeVM()).Remote(conn).Alert().High()

		toTn, err = auxlib.Copy(ctx, dst, conn)
		if err != nil {
			ev.Subject("高交互代理蜜罐关闭请求").Msg("程序名称:%s 代理关闭 发送:%d", p.Name(), toTn).E(err)
			p.pipe(ev)
		} else {
			ev.Subject("高交互代理蜜罐请求结束").Msg("程序名称: %s 发送到远程 发送:%d", p.Name(), toTn)
			p.pipe(ev)
		}
	})

	xEnv.Spawn(50, func() {
		defer func() {
			stop()
			dst.Close()
		}()
		var rev int64

		ev = audit.NewEvent("chameleon").From(p.cur.co.CodeVM()).Remote(conn).Alert().High()
		rev, err = auxlib.Copy(ctx, conn, dst)
		if err != nil {
			ev.Subject("高交互代理蜜罐关闭请求").Msg("程序名称:%s 接收远程失败:%d", p.Name(), rev).E(err)
			p.pipe(ev)

		} else {
			ev.Subject("高交互代理蜜罐请求结束").
				Msg("程序名称:%s 接收远程结束 数量:%d", p.Name(), rev)
			p.pipe(ev)
		}
	})

	return err
}
