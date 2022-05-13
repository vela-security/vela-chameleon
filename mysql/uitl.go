package mysql

import (
	"github.com/vela-security/vela-public/lua"
	"github.com/vela-security/vela-chameleon/mysql/sql"
)

func checkTableColName(L *lua.LState, val lua.LValue) string {
	if val.Type() != lua.LTString {
		L.RaiseError("table col Name must be string , got %s", val.Type().String())
		return ""
	}

	v := val.String()
	n := len(v)

	for i := 0; i < n; i++ {
		ch := v[i]
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}

		L.RaiseError("invalid Name")
	}
	return v
}

func checkTableColType(L *lua.LState, val lua.LValue) sql.Type {
	if val.Type() != lua.LTString {
		L.RaiseError("table col Name must be string , got %s", val.Type().String())
		return nil
	}

	switch val.String() {
	case "text":
		return sql.Text
	case "int":
		return sql.Int32
	case "float":
		return sql.Float32
	default:
		L.RaiseError("not found type")
		return nil
	}
}

func CheckDatabaseTable(L *lua.LState, val lua.LValue) *Table {
	if val.Type() != lua.LTAnyData {
		L.RaiseError("invalid type")
		return nil
	}

	t, ok := val.(*lua.AnyData).Data.(*Table)
	if !ok {
		L.RaiseError("invalid database table type")
		return nil
	}

	return t

}
