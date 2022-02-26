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

// Package testserver provides helpers to run a cockroach binary within tests.
// It automatically downloads the latest cockroach binary for your platform
// (Linux-amd64 and Darwin-amd64 only for now), or attempts to run "cockroach"
// from your PATH.
//
// To use, run as follows:
//   import "github.com/cockroachdb/cockroach-go/v2/testserver"
//   import "testing"
//   import "time"
//
//   func TestRunServer(t *testing.T) {
//      ts, err := testserver.NewTestServer()
//      if err != nil {
//        t.Fatal(err)
//      }
//      defer ts.Stop()
//
//      db, err := sql.Open("postgres", ts.PGURL().String())
//      if err != nil {
//        t.Fatal(err)
//      }
//    }
package testserver

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver/version"
	// Import postgres driver.
	_ "github.com/lib/pq"
)

var customBinaryFlag = flag.String("cockroach-binary", "", "Use specified cockroach binary")

const (
	stateNew = 1 + iota
	stateRunning
	stateStopped
	stateFailed
)

const (
	// First tenant ID to use is 2 since 1 belongs to the system tenant. Refer
	// to NewTenantServer for more information.
	firstTenantID = 2
)

// By default, we allocate 20% of available memory to the test server.
const defaultStoreMemSize = 0.2

const testserverMessagePrefix = "cockroach-go testserver"
const tenantserverMessagePrefix = "cockroach-go tenantserver"

// TestServer is a helper to run a real cockroach node.
type TestServer interface {
	// Start starts the server.
	Start() error
	// Stop stops the server and cleans up any associated resources.
	Stop()

	// Stdout returns the entire contents of the process' stdout.
	Stdout() string
	// Stdout returns the entire contents of the process' stderr.
	Stderr() string
	// PGURL returns the postgres connection URL to this server.
	PGURL() *url.URL
	// WaitForInit retries until a SQL connection is successfully established to
	// this server.
	WaitForInit() error
}

// testServerImpl is a TestServer implementation.
type testServerImpl struct {
	mu         sync.RWMutex
	version    *version.Version
	serverArgs testServerArgs
	state      int
	baseDir    string
	pgURL      struct {
		set chan struct{}
		u   *url.URL
		// The original URL is preserved here if we are using a custom password.
		// In that case, the one below uses client certificates, if secure (and
		// no password otherwise).
		orig url.URL
	}
	cmd              *exec.Cmd
	cmdArgs          []string
	stdout           string
	stderr           string
	stdoutBuf        logWriter
	stderrBuf        logWriter
	listeningURLFile string

	// curTenantID is used to allocate tenant IDs. Refer to NewTenantServer for
	// more information.
	curTenantID  int
	proxyAddr    string      // empty if no sql proxy running yet
	proxyProcess *os.Process // empty if no sql proxy running yet
}

// NewDBForTest creates a new CockroachDB TestServer instance and
// opens a SQL database connection to it. Returns a sql *DB instance and a
// shutdown function. The caller is responsible for executing the
// returned shutdown function on exit.
func NewDBForTest(t *testing.T, opts ...TestServerOpt) (*sql.DB, func()) {
	t.Helper()
	return NewDBForTestWithDatabase(t, "", opts...)
}

// NewDBForTestWithDatabase creates a new CockroachDB TestServer
// instance and opens a SQL database connection to it. If database is
// specified, the returned connection will explicitly connect to
// it. Returns a sql *DB instance a shutdown function. The caller is
// responsible for executing the returned shutdown function on exit.
func NewDBForTestWithDatabase(
	t *testing.T, database string, opts ...TestServerOpt,
) (*sql.DB, func()) {
	t.Helper()
	ts, err := NewTestServer(opts...)
	if err != nil {
		if errors.Is(err, errStoppedInMiddle) {
			// If the testserver is intentionally killed in the middle,
			// make sure it is stopped.
			return nil, func() {
				if ts != nil {
					ts.Stop()
				}
			}
		}
		t.Fatal(err)
	}
	url := ts.PGURL()
	if len(database) > 0 {
		url.Path = database
	}

	db, err := sql.Open("postgres", url.String())
	if err != nil {
		t.Fatalf("%s: %v", testserverMessagePrefix, err)
	}

	return db, func() {
		_ = db.Close()
		ts.Stop()
	}
}

// TestServerOpt is passed to NewTestServer.
type TestServerOpt func(args *testServerArgs)

type TestConfig struct {
	IsTest               bool
	StopDownloadInMiddle bool
}

type testServerArgs struct {
	secure       bool
	rootPW       string  // if nonempty, set as pw for root
	storeOnDisk  bool    // to save database in disk
	storeMemSize float64 // the proportion of available memory allocated to test server
	testConfig   TestConfig
	nonStableDB  bool
}

// SecureOpt is a TestServer option that can be passed to NewTestServer to
// enable secure mode.
func SecureOpt() TestServerOpt {
	return func(args *testServerArgs) {
		args.secure = true
	}
}

// StoreOnDiskOpt is a TestServer option that can be passed to NewTestServer
// to enable storing database in memory.
func StoreOnDiskOpt() TestServerOpt {
	return func(args *testServerArgs) {
		args.storeOnDisk = true
	}
}

// SetStoreMemSizeOpt is a TestServer option that can be passed to NewTestServer
// to set the proportion of available memory that is allocated
// to the test server.
func SetStoreMemSizeOpt(memSize float64) TestServerOpt {
	return func(args *testServerArgs) {
		if memSize > 0 {
			args.storeMemSize = memSize
		} else {
			args.storeMemSize = defaultStoreMemSize
		}
	}
}

// RootPasswordOpt is a TestServer option that, when passed to NewTestServer,
// sets the given password for the root user (and returns a URL using it from
// PGURL(). This avoids having to use client certs.
func RootPasswordOpt(pw string) TestServerOpt {
	return func(args *testServerArgs) {
		args.rootPW = pw
	}
}

// NonStableDbOpt is a TestServer option that can be passed to NewTestServer to
// download the latest beta version of CRDB, but not necessary a stable one.
func NonStableDbOpt() TestServerOpt {
	return func(args *testServerArgs) {
		args.nonStableDB = true
	}
}

// StopDownloadInMiddleOpt is a TestServer option used only in testing.
// It is used to test the flock over downloaded CRDB binary.
// It should not be used in production.
func StopDownloadInMiddleOpt() TestServerOpt {
	return func(args *testServerArgs) {
		tc := TestConfig{IsTest: true, StopDownloadInMiddle: true}
		args.testConfig = tc
	}
}

const (
	logsDirName  = "logs"
	certsDirName = "certs"
)

var errStoppedInMiddle = errors.New("download stopped in middle")

// NewTestServer creates a new TestServer and starts it.
// It also waits until the server is ready to accept clients,
// so it safe to connect to the server returned by this function right away.
// The cockroach binary for your OS and ARCH is downloaded automatically.
// If the download fails, we attempt just call "cockroach", hoping it is
// found in your path.
func NewTestServer(opts ...TestServerOpt) (TestServer, error) {
	serverArgs := &testServerArgs{}
	serverArgs.storeMemSize = defaultStoreMemSize
	for _, applyOptToArgs := range opts {
		applyOptToArgs(serverArgs)
	}

	var cockroachBinary string

	if len(*customBinaryFlag) > 0 {
		cockroachBinary = *customBinaryFlag
	} else if customBinaryEnv := os.Getenv("COCKROACH_BINARY"); customBinaryEnv != "" {
		cockroachBinary = customBinaryEnv
	}

	var err error
	if cockroachBinary != "" {
		log.Printf("Using custom cockroach binary: %s", cockroachBinary)
	} else {
		cockroachBinary, err = downloadBinary(&serverArgs.testConfig, serverArgs.nonStableDB)
		if err != nil {
			if errors.Is(err, errStoppedInMiddle) {
				// If the testserver is intentionally killed in the middle of downloading,
				// return error.
				return nil, err
			}
			log.Printf("%s: Failed to fetch latest binary: %v attempting to use cockroach binary from your PATH", testserverMessagePrefix, err)
			cockroachBinary = "cockroach"
		} else {
			log.Printf("Using automatically-downloaded binary: %s", cockroachBinary)
		}
	}

	// Force "/tmp/" so avoid OSX's really long temp directory names
	// which get us over the socket filename length limit.
	baseDir, err := ioutil.TempDir("/tmp", "cockroach-testserver")
	if err != nil {
		return nil, fmt.Errorf("%s: could not create temp directory: %w",
			testserverMessagePrefix, err)
	}

	mkDir := func(name string) (string, error) {
		path := filepath.Join(baseDir, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", fmt.Errorf("%s: could not create %s directory: %s: %w",
				testserverMessagePrefix, name, path, err)
		}
		return path, nil
	}
	// TODO(janexing): Make sure the log is written to logDir instead of shown in console.
	// Should be done once issue #109 is solved:
	// https://github.com/cockroachdb/cockroach-go/issues/109
	logDir, err := mkDir(logsDirName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", testserverMessagePrefix, err)
	}
	certsDir, err := mkDir(certsDirName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", testserverMessagePrefix, err)
	}

	listeningURLFile := filepath.Join(baseDir, "listen-url")

	secureOpt := "--insecure"
	if serverArgs.secure {
		// Create certificates.
		certArgs := []string{
			"--certs-dir=" + certsDir,
			"--ca-key=" + filepath.Join(certsDir, "ca.key"),
		}
		for _, args := range [][]string{
			// Create the CA cert and key pair.
			{"cert", "create-ca"},
			// Create cert and key pair for the cockroach node.
			{"cert", "create-node", "localhost"},
			// Create cert and key pair for the root user (SQL client).
			{"cert", "create-client", "root", "--also-generate-pkcs8-key"},
		} {
			createCertCmd := exec.Command(cockroachBinary, append(args, certArgs...)...)
			log.Printf("%s executing: %s", testserverMessagePrefix, createCertCmd)
			if err := createCertCmd.Run(); err != nil {
				return nil, fmt.Errorf("%s command %s failed: %w", testserverMessagePrefix, createCertCmd, err)
			}
		}
		secureOpt = "--certs-dir=" + certsDir
	}

	// v19.1 and earlier do not have the `start-single-node` subcommand,
	// so use `start` for those versions.
	// TODO(rafi): Remove the version check and `start` once we stop testing 19.1.
	versionCmd := exec.Command(cockroachBinary, "version")
	versionOutput, err := versionCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s command %s failed: %w", testserverMessagePrefix, versionCmd, err)
	}
	reader := bufio.NewReader(bytes.NewReader(versionOutput))
	versionLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("%s failed to read version: %w", testserverMessagePrefix, err)
	}
	versionLineTokens := strings.Fields(versionLine)
	v, err := version.Parse(versionLineTokens[2])
	if err != nil {
		return nil, fmt.Errorf("%s failed to parse version: %w", testserverMessagePrefix, err)
	}
	startCmd := "start-single-node"
	if !v.AtLeast(version.MustParse("v19.2.0-alpha")) {
		startCmd = "start"
	}

	var storeArg string
	if serverArgs.storeOnDisk {
		storeArg = "--store=path=" + baseDir
	} else {
		storeArg = fmt.Sprintf("--store=type=mem,size=%.2f", serverArgs.storeMemSize)
	}

	args := []string{
		cockroachBinary,
		startCmd,
		"--logtostderr",
		secureOpt,
		"--host=localhost",
		"--port=0",
		"--http-port=0",
		storeArg,
		"--listening-url-file=" + listeningURLFile,
	}

	ts := &testServerImpl{
		serverArgs:       *serverArgs,
		version:          v,
		state:            stateNew,
		baseDir:          baseDir,
		cmdArgs:          args,
		stdout:           filepath.Join(logDir, "cockroach.stdout"),
		stderr:           filepath.Join(logDir, "cockroach.stderr"),
		listeningURLFile: listeningURLFile,
		curTenantID:      firstTenantID,
	}
	ts.pgURL.set = make(chan struct{})

	if err := ts.Start(); err != nil {
		return nil, fmt.Errorf("%s Start failed: %w", testserverMessagePrefix, err)
	}

	if ts.PGURL() == nil {
		return nil, fmt.Errorf("%s: url not found", testserverMessagePrefix)
	}

	if err := ts.WaitForInit(); err != nil {
		return nil, fmt.Errorf("%s WaitForInit failed: %w", testserverMessagePrefix, err)
	}

	return ts, nil
}

// Stdout returns the entire contents of the process' stdout.
func (ts *testServerImpl) Stdout() string {
	return ts.stdoutBuf.String()
}

// Stderr returns the entire contents of the process' stderr.
func (ts *testServerImpl) Stderr() string {
	return ts.stderrBuf.String()
}

// PGURL returns the postgres connection URL to reach the started
// cockroach node.
//
// It blocks until the network URL is determined and does not timeout,
// relying instead on test timeouts.
func (ts *testServerImpl) PGURL() *url.URL {
	<-ts.pgURL.set
	return ts.pgURL.u
}

func (ts *testServerImpl) setPGURL(u *url.URL) {
	ts.pgURL.u = u
	close(ts.pgURL.set)
}

// WaitForInit retries until a connection is successfully established.
func (ts *testServerImpl) WaitForInit() error {
	var err error
	db, err := sql.Open("postgres", ts.PGURL().String())
	if err != nil {
		return err
	}
	defer db.Close()
	for i := 0; i < 50; i++ {
		if _, err = db.Query("SHOW DATABASES"); err == nil {
			return err
		}
		log.Printf("%s: WaitForInit: Trying again after error: %v", testserverMessagePrefix, err)
		time.Sleep(time.Millisecond * 100)
	}
	return err
}

func (ts *testServerImpl) pollListeningURLFile() error {
	var data []byte
	for {
		ts.mu.Lock()
		state := ts.state
		ts.mu.Unlock()
		if state != stateRunning {
			return fmt.Errorf("server stopped or crashed before listening URL file was available")
		}

		var err error
		data, err = ioutil.ReadFile(ts.listeningURLFile)
		if err == nil {
			break
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("unexpected error while reading listening URL file: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	u, err := url.Parse(string(bytes.TrimSpace(data)))
	if err != nil {
		return fmt.Errorf("failed to parse SQL URL: %w", err)
	}
	ts.pgURL.orig = *u
	if pw := ts.serverArgs.rootPW; pw != "" {
		db, err := sql.Open("postgres", u.String())
		if err != nil {
			return err
		}
		defer db.Close()
		if _, err := db.Exec(`ALTER USER root WITH PASSWORD $1`, pw); err != nil {
			return err
		}

		v := u.Query()
		v.Del("sslkey")
		v.Del("sslcert")
		u.RawQuery = v.Encode()
		u.User = url.UserPassword("root", pw)
	}
	ts.setPGURL(u)

	return nil
}

// Start runs the process, returning an error on any problems,
// including being unable to start, but not unexpected failure.
// It should only be called once in the lifetime of a TestServer object.
// If the server is already running, this function is a no-op.
// If the server stopped or failed, please don't use ts.Start()
// to restart a testserver, but use NewTestServer().
func (ts *testServerImpl) Start() error {
	ts.mu.Lock()
	if ts.state != stateNew {
		ts.mu.Unlock()
		switch ts.state {
		case stateRunning:
			return nil // No-op if server is already running.
		case stateStopped, stateFailed:
			// Start() can only be called once.
			return errors.New(
				"`Start()` cannot be used to restart a stopped or failed server. " +
					"Please use NewTestServer()")
		}
	}
	ts.state = stateRunning
	ts.mu.Unlock()

	ts.cmd = exec.Command(ts.cmdArgs[0], ts.cmdArgs[1:]...)
	ts.cmd.Env = []string{
		"COCKROACH_MAX_OFFSET=1ns",
		"COCKROACH_TRUST_CLIENT_PROVIDED_SQL_REMOTE_ADDR=true",
	}

	if len(ts.stdout) > 0 {
		wr, err := newFileLogWriter(ts.stdout)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", ts.stdout, err)
		}
		ts.stdoutBuf = wr
	}
	ts.cmd.Stdout = ts.stdoutBuf

	if len(ts.stderr) > 0 {
		wr, err := newFileLogWriter(ts.stderr)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", ts.stderr, err)
		}
		ts.stderrBuf = wr
	}
	ts.cmd.Stderr = ts.stderrBuf

	for k, v := range defaultEnv() {
		ts.cmd.Env = append(ts.cmd.Env, k+"="+v)
	}

	log.Printf("executing: %s", ts.cmd)
	err := ts.cmd.Start()
	if ts.cmd.Process != nil {
		log.Printf("process %d started: %s", ts.cmd.Process.Pid, strings.Join(ts.cmdArgs, " "))
	}
	if err != nil {
		log.Print(err.Error())
		if err := ts.stdoutBuf.Close(); err != nil {
			log.Printf("%s: failed to close stdout: %v", testserverMessagePrefix, err)
		}
		if err := ts.stderrBuf.Close(); err != nil {
			log.Printf("%s: failed to close stderr: %v", testserverMessagePrefix, err)
		}

		ts.mu.Lock()
		ts.state = stateFailed
		ts.mu.Unlock()

		return fmt.Errorf("command %s failed: %w", ts.cmd, err)
	}

	go func() {
		err := ts.cmd.Wait()

		if closeErr := ts.stdoutBuf.Close(); closeErr != nil {
			log.Printf("%s: failed to close stdout: %v", testserverMessagePrefix, closeErr)
		}
		if closeErr := ts.stderrBuf.Close(); closeErr != nil {
			log.Printf("%s: failed to close stderr: %v", testserverMessagePrefix, closeErr)
		}

		ps := ts.cmd.ProcessState
		sy := ps.Sys().(syscall.WaitStatus)

		log.Printf("%s: command %s exited with status %d: %v",
			testserverMessagePrefix,
			ts.cmd,
			sy.ExitStatus(),
			err)
		log.Printf("%s process state: %s", testserverMessagePrefix, ps.String())

		ts.mu.Lock()
		if sy.ExitStatus() == 0 {
			ts.state = stateStopped
		} else {
			ts.state = stateFailed
		}
		ts.mu.Unlock()
	}()

	if ts.pgURL.u == nil {
		go func() {
			if err := ts.pollListeningURLFile(); err != nil {
				log.Printf("%s failed to poll listening URL file: %v", testserverMessagePrefix, err)
				close(ts.pgURL.set)
				ts.Stop()
			}
		}()
	}

	return nil
}

// Stop kills the process if it is still running and cleans its directory.
// It should only be called once in the lifetime of a TestServer object.
// Logs fatal if the process has already failed.
func (ts *testServerImpl) Stop() {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if ts.state == stateNew {
		log.Fatalf("%s: Stop() called, but Start() was never called", testserverMessagePrefix)
	}
	if ts.state == stateFailed {
		log.Fatalf("%s: Stop() called, but process exited unexpectedly. Stdout:\n%s\nStderr:\n%s\n",
			testserverMessagePrefix,
			ts.Stdout(),
			ts.Stderr())
		return
	}

	if ts.state != stateStopped {
		// Only call kill if not running. It could have exited properly.
		_ = ts.cmd.Process.Kill()

		if p := ts.proxyProcess; p != nil {
			_ = p.Kill()
		}
	}

	// Only cleanup on intentional stops.
	_ = os.RemoveAll(ts.baseDir)
}

type logWriter interface {
	Write(p []byte) (n int, err error)
	String() string
	Len() int64
	Close() error
}

type fileLogWriter struct {
	filename string
	file     *os.File
}

func newFileLogWriter(file string) (*fileLogWriter, error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	return &fileLogWriter{
		filename: file,
		file:     f,
	}, nil
}

func (w fileLogWriter) Close() error {
	return w.file.Close()
}

func (w fileLogWriter) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

func (w fileLogWriter) String() string {
	b, err := ioutil.ReadFile(w.filename)
	if err == nil {
		return string(b)
	}
	return ""
}

func (w fileLogWriter) Len() int64 {
	s, err := os.Stat(w.filename)
	if err == nil {
		return s.Size()
	}
	return 0
}

func defaultEnv() map[string]string {
	vars := map[string]string{}
	u, err := user.Current()
	if err == nil {
		if _, ok := vars["USER"]; !ok {
			vars["USER"] = u.Username
		}
		if _, ok := vars["UID"]; !ok {
			vars["UID"] = u.Uid
		}
		if _, ok := vars["GID"]; !ok {
			vars["GID"] = u.Gid
		}
		if _, ok := vars["HOME"]; !ok {
			vars["HOME"] = u.HomeDir
		}
	}
	if _, ok := vars["PATH"]; !ok {
		vars["PATH"] = os.Getenv("PATH")
	}
	return vars
}
