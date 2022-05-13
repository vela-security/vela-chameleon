package mysql

import (
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-chameleon/mysql/memory"
	"github.com/vela-security/vela-chameleon/mysql/sql"
)

type Table struct {
	n    int //列数
	name string
	tab  *memory.Table
	meta lua.UserKV
}

func newTable(L *lua.LState) *Table {
	n := L.GetTop()
	if n < 2 {
		L.RaiseError("not found table schema")
		return nil
	}

	name := L.CheckString(1)
	ssa := make([]*sql.Column, n-1)
	for i := 2; i <= n; i++ {
		_tab := L.CheckTable(i)
		_name := checkTableColName(L, _tab.RawGetString("name"))
		_type := checkTableColType(L, _tab.RawGetString("type"))
		_null := _tab.CheckBool("null", true)
		_src := name
		ssa[i-2] = &sql.Column{
			Name:     _name,
			Type:     _type,
			Nullable: _null,
			Source:   _src,
		}
	}
	t := &Table{name: name, n: n - 1}
	t.tab = memory.NewTable(name, sql.Schema(ssa))
	t.initMeta()

	return t
}

func (t *Table) initMeta() {
	t.meta = lua.NewUserKV()
	t.meta.Set("insert", lua.NewFunction(t.insert))
}

func (t *Table) insert(L *lua.LState) int {
	n := L.GetTop()
	if n != t.n {
		L.RaiseError("%s have %d col , got %d value", t.n, n)
		return 0
	}

	var row []interface{}
	for i := 1; i <= n; i++ {
		val := L.Get(i)
		switch val.Type() {
		case lua.LTString:
			row = append(row, val.String())
		case lua.LTNumber:
			row = append(row, int(val.(lua.LNumber)))
		default:
			L.RaiseError("inset invalid value")
			return 0
		}
	}

	ctx := sql.NewEmptyContext()
	e := t.tab.Insert(ctx, sql.NewRow(row...))
	if e != nil {
		L.RaiseError("%v", e)
		return 0
	}

	return 0

}

func (t *Table) Index(L *lua.LState, key string) lua.LValue {
	return t.meta.Get(key)
}
