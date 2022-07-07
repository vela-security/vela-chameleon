package mysql

import (
	"github.com/vela-security/vela-chameleon/mysql/auth"
	"github.com/vela-security/vela-chameleon/mysql/sql"
	risk "github.com/vela-security/vela-risk"
	"time"
)

type Audit struct {
	CodeVM func() string
}

func (a *Audit) Authentication(user, addr string, err error) {
	ev := risk.HoneyPot()
	ev.Remote(addr)
	ev.From(a.CodeVM())
	ev.Leve(risk.HIGH)

	if err == nil {
		ev.Subjectf("mysql蜜罐认证成功")
		ev.Payloadf("user:%s", user)
	} else {
		ev.Subjectf("mysql蜜罐认证失败")
		ev.Payloadf("user:%s err:%v", user, err)
	}
	ev.Send()
}

func (a *Audit) Authorization(ctx *sql.Context, p auth.Permission, err error) {
	//	fmt.Printf("authO: %s %s %v\n" ,  ctx.Session , ctx.Client().Address, p)

}
func (a *Audit) Query(ctx *sql.Context, d time.Duration, err error) {
	//  "user":          ctx.Client().User,
	//	"query":         ctx.Query(),
	//	"address":       ctx.Client().Address,
	//	"connection_id": ctx.Session.ID(),
	//	"pid":           ctx.Pid(),
	//	"success":       true,

	//fmt.Printf("Query: %s %s %s %v %s\n" , d , ctx.Session , ctx.Client().Address, ctx.Query())
}
