/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2019 Hajime Nakagami

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*******************************************************************************/

package firebirdsql

import (
	"testing"
)

func TestDSNParse(t *testing.T) {
	var testDSNs = []struct {
		dsn            string
		addr           string
		dbName         string
		user           string
		passwd         string
		role           string
		authPluginName string
		wireCrypt      string
	}{
		{"user:password@localhost:3000/dbname", "localhost:3000", "dbname", "user", "password", "", "Srp", "true"},
		{"user:password@localhost/dbname", "localhost:3050", "dbname", "user", "password", "", "Srp", "true"},
		{"user:password@localhost/dir/dbname", "localhost:3050", "/dir/dbname", "user", "password", "", "Srp", "true"},
		{"user:password@localhost/c:\\fbdata\\database.fdb", "localhost:3050", "c:\\fbdata\\database.fdb", "user", "password", "", "Srp", "true"},
		{"user:password@localhost/c:/fbdata/database.fdb", "localhost:3050", "c:/fbdata/database.fdb", "user", "password", "", "Srp", "true"},
		{"user:password@localhost/dbname?role=role", "localhost:3050", "dbname", "user", "password", "role", "Srp", "true"},
		{"user:password@localhost:3000/c:/fbdata/database.fdb?role=role&wire_crypt=false", "localhost:3000", "c:/fbdata/database.fdb", "user", "password", "role", "Srp", "false"},
		{"firebird://user:password@localhost:3000/dbname", "localhost:3000", "dbname", "user", "password", "", "Srp", "true"},
		{"firebird://user:%21p%40ssword%3F@localhost:3050/dbname", "localhost:3050", "dbname", "user", "!p@ssword?", "", "Srp", "true"},
	}

	for _, d := range testDSNs {
		dsn, err := parseDSN(d.dsn)
		if dsn == nil {
			t.Fatal("parse DSN fail. firebirdDsn is nil.")
		}
		if err != nil {
			t.Fatal(err)
		}
		if dsn.addr != d.addr {
			t.Errorf("parse DSN fail:%s(%s != %s)", d.dsn, dsn.addr, d.addr)
		}
		if dsn.dbName != d.dbName {
			t.Errorf("parse DSN fail:%s(%s != %s)", d.dsn, dsn.dbName, d.dbName)
		}
		if dsn.user != d.user {
			t.Errorf("parse DSN fail:%s(%s != %s)", d.dsn, dsn.user, d.user)
		}
		if dsn.passwd != d.passwd {
			t.Errorf("parse DSN fail:%s(%s != %s)", d.dsn, dsn.passwd, d.passwd)
		}
		if dsn.options["role"] != d.role {
			t.Errorf("parse DSN fail:%s(%s != %s)", d.dsn, dsn.options["role"], d.role)
		}
		if dsn.options["auth_plugin_name"] != d.authPluginName {
			t.Errorf("parse DSN fail:%s(%s != %s)", d.dsn, dsn.options["auth_plugin_name"], d.authPluginName)
		}
		if dsn.options["wire_crypt"] != d.wireCrypt {
			t.Errorf("parse DSN fail:%s(%v != %v)", d.dsn, dsn.options["wire_crypt"], d.wireCrypt)
		}
	}

	_, err := parseDSN("something wrong")
	if err == nil {
		t.Fatalf("Error Not occured")
	}
	_, err = parseDSN("SomethingWrongConnectionString")
	if err == nil {
		t.Fatalf("Error Not occured")
	}

}
