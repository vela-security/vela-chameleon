package mysql

import (
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-chameleon/mysql/memory"
)

type EngineDB struct {
	obj  *memory.Database
	meta lua.UserKV
}

func newEngineDB(name string) *EngineDB {
	edb := &EngineDB{}
	edb.obj = memory.NewDatabase(name)
	edb.meta = lua.NewUserKV()
	edb.initMeta()
	return edb
}

func (edb *EngineDB) addTable(L *lua.LState) int {
	n := L.GetTop()
	if n <= 0 {
		return 0
	}

	for i := 1; i <= n; i++ {
		val := L.Get(i)
		t := CheckDatabaseTable(L, val)
		edb.obj.AddTable(t.name, t.tab)
	}
	return 0
}

func (edb *EngineDB) initMeta() {
	edb.meta.Set("add_table", lua.NewFunction(edb.addTable))
}

func (edb *EngineDB) Index(L *lua.LState, key string) lua.LValue {
	return edb.meta.Get(key)
}
