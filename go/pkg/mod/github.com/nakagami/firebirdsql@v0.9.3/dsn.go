/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2020 Hajime Nakagami

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
	"errors"
	"net/url"
	"strings"
)

type firebirdDsn struct {
	addr    string
	dbName  string
	user    string
	passwd  string
	options map[string]string
}

var ErrDsnUserUnknown = errors.New("User unknown")

func newFirebirdDsn() *firebirdDsn {
	return &firebirdDsn{options: make(map[string]string)}
}

func parseDSN(dsns string) (*firebirdDsn, error) {

	dsn := newFirebirdDsn()

	if !strings.HasPrefix(dsns, "firebird://") {
		dsns = "firebird://" + dsns
	}
	u, err := url.Parse(dsns)
	if err != nil {
		return nil, err
	}
	if u.User == nil {
		return nil, ErrDsnUserUnknown
	}
	dsn.user = u.User.Username()
	dsn.passwd, _ = u.User.Password()
	dsn.addr = u.Host
	if !strings.ContainsRune(dsn.addr, ':') {
		dsn.addr += ":3050"
	}
	dsn.dbName = u.Path
	if !strings.ContainsRune(dsn.dbName[1:], '/') {
		dsn.dbName = dsn.dbName[1:]
	}

	//Windows Path
	if strings.ContainsRune(dsn.dbName[2:], ':') {
		dsn.dbName = dsn.dbName[1:]
	}

	m, _ := url.ParseQuery(u.RawQuery)

	var default_options = map[string]string{
		"auth_plugin_name":     "Srp",
		"charset":              "UTF8",
		"column_name_to_lower": "false",
		"role":                 "",
		"timezone":             "",
		"wire_crypt":           "true",
	}

	for k, v := range default_options {
		values, ok := m[k]
		if ok {
			dsn.options[k] = values[0]
		} else {
			dsn.options[k] = v
		}
	}

	return dsn, nil
}
