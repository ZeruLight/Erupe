// Copyright 2020 The Cockroach Authors.
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

package testserver

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
)

func (ts *testServerImpl) isTenant() bool {
	// ts.curTenantID is initialized to firstTenantID in system tenant servers.
	// An uninitialized ts.curTenantID indicates that this TestServer is a
	// tenant.
	return ts.curTenantID < firstTenantID
}

// NewTenantServer creates and returns a new SQL tenant pointed at the receiver,
// which acts as a KV server.
// The SQL tenant is responsible for all SQL processing and does not store any
// physical KV pairs. It issues KV RPCs to the receiver. The idea is to be able
// to create multiple SQL tenants each with an exclusive keyspace accessed
// through the KV server.
// WARNING: This functionality is internal and experimental and subject to
// change. See cockroach mt start-sql --help.
// NOTE: To use this, a caller must first define an interface that includes
// NewTenantServer, and subsequently cast the TestServer obtained from
// NewTestServer to this interface. Refer to the tests for an example.
func (ts *testServerImpl) NewTenantServer() (TestServer, error) {
	tenantID, err := func() (int, error) {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		if ts.state != stateRunning {
			return 0, errors.New("TestServer must be running before NewTenantServer may be called")
		}
		if ts.isTenant() {
			return 0, errors.New("cannot call NewTenantServer on a tenant")
		}
		tenantID := ts.curTenantID
		ts.curTenantID++
		return tenantID, nil
	}()
	if err != nil {
		return nil, err
	}

	secureFlag := "--insecure"
	if ts.serverArgs.secure {
		secureFlag = "--certs-dir=" + filepath.Join(ts.baseDir, "certs")
	}

	// Create a new tenant.
	pgURL := ts.PGURL()
	if pgURL == nil {
		return nil, errors.New("url not found")
	}
	db, err := sql.Open("postgres", pgURL.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := ts.WaitForInit(db); err != nil {
		return nil, err
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT crdb_internal.create_tenant(%d)", tenantID)); err != nil {
		return nil, err
	}

	// TODO(asubiotto): We should pass ":0" as the sql addr to push port
	//  selection to the cockroach mt start-sql command. However, that requires
	//  that the mt start-sql command supports --listening-url-file so that this
	//  test harness can subsequently read the postgres url. The current
	//  approach is to do our best to find a free port and use that.
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	// Use localhost because of certificate validation issues otherwise
	// (something about IP SANs).
	sqlAddr := "localhost:" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		return nil, err
	}

	args := []string{
		ts.cmdArgs[0],
		"mt",
		"start-sql",
		secureFlag,
		"--logtostderr",
		fmt.Sprintf("--tenant-id=%d", tenantID),
		"--kv-addrs=" + pgURL.Host,
		"--sql-addr=" + sqlAddr,
	}

	tenant := &testServerImpl{
		serverArgs: ts.serverArgs,
		state:      stateNew,
		baseDir:    ts.baseDir,
		cmdArgs:    args,
		stdout:     filepath.Join(ts.baseDir, logsDirName, fmt.Sprintf("cockroach.tenant.%d.stdout", tenantID)),
		stderr:     filepath.Join(ts.baseDir, logsDirName, fmt.Sprintf("cockroach.tenant.%d.stderr", tenantID)),
		// TODO(asubiotto): Specify listeningURLFile once we support dynamic
		//  ports.
		listeningURLFile: "",
	}
	tenant.pgURL.set = make(chan struct{})
	// Copy and overwrite the TestServer's url host:port.
	tenantURL := *pgURL
	tenantURL.Host = sqlAddr
	tenant.setPGURL(&tenantURL)
	return tenant, nil
}
