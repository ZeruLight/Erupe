// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package testserver_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach-go/testserver"
)

func TestRunServer(t *testing.T) {
	for _, tc := range []struct {
		name          string
		instantiation func() (*sql.DB, func())
	}{
		{
			name:          "Insecure",
			instantiation: func() (*sql.DB, func()) { return testserver.NewDBForTest(t) },
		},
		{
			name:          "Secure",
			instantiation: func() (*sql.DB, func()) { return testserver.NewDBForTest(t, testserver.SecureOpt()) },
		},
		{
			name:          "InsecureTenant",
			instantiation: func() (*sql.DB, func()) { return newTenantDBForTest(t, false /* secure */) },
		},
		{
			name:          "SecureTenant",
			instantiation: func() (*sql.DB, func()) { return newTenantDBForTest(t, true /* secure */) },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, stop := tc.instantiation()
			defer stop()
			if _, err := db.Exec("SELECT 1"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPGURLWhitespace(t *testing.T) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Start(); err != nil {
		t.Fatal(err)
	}
	url := ts.PGURL().String()
	if trimmed := strings.TrimSpace(url); url != trimmed {
		t.Errorf("unexpected whitespace in server URL: %q", url)
	}
}

// tenantInterface is defined in order to use tenant-related methods on the
// TestServer.
type tenantInterface interface {
	NewTenantServer() (testserver.TestServer, error)
}

// newTenantDBForTest is a testing helper function that starts a TestServer
// process and a SQL tenant process pointed at this TestServer. A sql connection
// to the tenant and a cleanup function are returned.
func newTenantDBForTest(t *testing.T, secure bool) (*sql.DB, func()) {
	t.Helper()
	var (
		ts  testserver.TestServer
		err error
	)
	if secure {
		ts, err = testserver.NewTestServer(testserver.SecureOpt())
	} else {
		ts, err = testserver.NewTestServer()
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Start(); err != nil {
		t.Fatal(err)
	}
	tenant, err := ts.(tenantInterface).NewTenantServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := tenant.Start(); err != nil {
		t.Fatal(err)
	}
	url := tenant.PGURL()
	if url == nil {
		t.Fatal("postgres url not found")
	}
	db, err := sql.Open("postgres", url.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := tenant.WaitForInit(db); err != nil {
		t.Fatal(err)
	}
	return db, func() {
		_ = db.Close()
		tenant.Stop()
		ts.Stop()
	}
}

func TestTenant(t *testing.T) {
	db, stop := newTenantDBForTest(t, false /* secure */)
	defer stop()
	if _, err := db.Exec("SELECT 1"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec("SELECT crdb_internal.create_tenant(123)"); err == nil {
		t.Fatal("expected an error when creating a tenant since secondary tenants should not be able to do so")
	}
}
