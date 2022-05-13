package mysql

import (
	"context"
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-chameleon/mysql/engine"
	"github.com/vela-security/vela-chameleon/mysql/server"
	"github.com/vela-security/vela-chameleon/mysql/sql/information_schema"
	"reflect"
)

var TGoMySQL = reflect.TypeOf((*GoMysql)(nil)).String()

type GoMysql struct {
	lua.ProcEx

	cfg    *config
	ser    *server.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func newGoMysql(cfg *config) *GoMysql {
	m := &GoMysql{cfg: cfg}
	m.V(lua.PTInit, TGoMySQL)
	return m
}

func (m *GoMysql) Name() string {
	return m.cfg.Name
}

func (m *GoMysql) Start() error {
	eg := engine.NewDefault()
	eg.AddDatabase(m.cfg.Database.obj)
	eg.AddDatabase(information_schema.NewInformationSchemaDatabase(eg.Catalog))

	s, err := server.NewDefaultServer(m.cfg.toSerCfg(), eg)
	if err != nil {
		return err
	}

	m.ser = s
	m.ser.CodeVM = func() string {
		return m.cfg.CodeVM
	}
	xEnv.Spawn(3, func() { err = s.Start() })

	m.ctx, m.cancel = context.WithCancel(context.Background())
	xEnv.Errorf("%s %s start succeed", m.Name(), m.Type())
	return nil
}

func (m *GoMysql) Close() error {
	m.cancel()
	return m.ser.Close()
}
