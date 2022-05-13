// Copyright 2020-2021 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"net"
	"time"

	"github.com/vela-security/vela-chameleon/mysql/sql"
	"github.com/vela-security/vela-chameleon/vitess/go/mysql"
)

// AuditMethod is called to log the vela-audit trail of actions.
type AuditMethod interface {
	// Authentication logs an authentication vela-event.
	Authentication(user, address string, err error)
	// Authorization logs an authorization vela-event.
	Authorization(ctx *sql.Context, p Permission, err error)
	// Query logs a query execution.
	Query(ctx *sql.Context, d time.Duration, err error)
}

// MysqlAudit wraps mysql.AuthServer to emit vela-audit trails.
type MysqlAudit struct {
	mysql.AuthServer
	audit AuditMethod
}

// ValidateHash sends authentication calls to an AuditMethod.
func (m *MysqlAudit) ValidateHash(
	salt []byte,
	user string,
	resp []byte,
	addr net.Addr,
) (mysql.Getter, error) {
	getter, err := m.AuthServer.ValidateHash(salt, user, resp, addr)
	m.audit.Authentication(user, addr.String(), err)

	return getter, err
}

// NewAudit creates a wrapped Auth that sends vela-audit trails to the specified
// method.
func NewAudit(auth Auth, method AuditMethod) Auth {
	return &Audit{
		auth:   auth,
		method: method,
	}
}

// Audit is an Auth method proxy that sends vela-audit trails to the specified
// AuditMethod.
type Audit struct {
	auth   Auth
	method AuditMethod
}

// Mysql implements Auth interface.
func (a *Audit) Mysql() mysql.AuthServer {
	return &MysqlAudit{
		AuthServer: a.auth.Mysql(),
		audit:      a.method,
	}
}

// Allowed implements Auth interface.
func (a *Audit) Allowed(ctx *sql.Context, permission Permission) error {
	err := a.auth.Allowed(ctx, permission)
	a.method.Authorization(ctx, permission, err)

	return err
}

// Query implements AuditQuery interface.
func (a *Audit) Query(ctx *sql.Context, d time.Duration, err error) {
	if q, ok := a.auth.(*Audit); ok {
		q.Query(ctx, d, err)
	}

	a.method.Query(ctx, d, err)
}
