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
//   import "github.com/cockroachdb/cockroach-go/testserver"
//   import "testing"
//   import "time"
//
//   func TestRunServer(t *testing.T) {
//      ts, err := testserver.NewTestServer()
//      if err != nil {
//        t.Fatal(err)
//      }
//      if err := ts.Start(); err != nil {
//        t.Fatal(err)
//      }
//      defer ts.Stop()
//
//      url := ts.PGURL()
//      if url == nil {
//        t.Fatalf("url not found")
//      }
//      t.Logf("URL: %s", url.String())
//
//      db, err := sql.Open("postgres", url.String())
//      if err != nil {
//        t.Fatal(err)
//      }
//    }
package testserver

import (
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

	// Import postgres driver.
	_ "github.com/lib/pq"
)

var customBinary = flag.String("cockroach-binary", "", "Use specified cockroach binary")

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
	WaitForInit(db *sql.DB) error
}

// testServerImpl is a TestServer implementation.
type testServerImpl struct {
	mu         sync.RWMutex
	serverArgs testServerArgs
	state      int
	baseDir    string
	pgURL      struct {
		set chan struct{}
		u   *url.URL
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
	curTenantID int
}

// NewDBForTest creates a new CockroachDB TestServer instance and
// opens a SQL database connection to it. Returns a sql *DB instance and a
// shutdown function. The caller is responsible for executing the
// returned shutdown function on exit.
func NewDBForTest(t *testing.T, opts ...testServerOpt) (*sql.DB, func()) {
	t.Helper()
	return NewDBForTestWithDatabase(t, "", opts...)
}

// NewDBForTestWithDatabase creates a new CockroachDB TestServer
// instance and opens a SQL database connection to it. If database is
// specified, the returned connection will explicitly connect to
// it. Returns a sql *DB instance a shutdown function. The caller is
// responsible for executing the returned shutdown function on exit.
func NewDBForTestWithDatabase(t *testing.T, database string, opts ...testServerOpt) (*sql.DB, func()) {
	t.Helper()
	ts, err := NewTestServer(opts...)
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Start(); err != nil {
		t.Fatal(err)
	}

	url := ts.PGURL()
	if url == nil {
		t.Fatalf("url not found")
	}
	if len(database) > 0 {
		url.Path = database
	}

	db, err := sql.Open("postgres", url.String())
	if err != nil {
		t.Fatal(err)
	}

	if err := ts.WaitForInit(db); err != nil {
		t.Fatal(err)
	}

	return db, func() {
		_ = db.Close()
		ts.Stop()
	}
}

type testServerOpt func(args *testServerArgs)

type testServerArgs struct {
	secure bool
}

// SecureOpt is a TestServer option that can be passed to NewTestServer to
// enable secure mode.
func SecureOpt() testServerOpt {
	return func(args *testServerArgs) {
		args.secure = true
	}
}

const (
	logsDirName  = "logs"
	certsDirName = "certs"
)

// NewTestServer creates a new TestServer, but does not start it.
// The cockroach binary for your OS and ARCH is downloaded automatically.
// If the download fails, we attempt just call "cockroach", hoping it is
// found in your path.
func NewTestServer(opts ...testServerOpt) (TestServer, error) {
	serverArgs := &testServerArgs{}
	for _, applyOptToArgs := range opts {
		applyOptToArgs(serverArgs)
	}

	var cockroachBinary string
	var err error
	if len(*customBinary) > 0 {
		cockroachBinary = *customBinary
		log.Printf("Using custom cockroach binary: %s", cockroachBinary)
	} else if cockroachBinary, err = downloadLatestBinary(); err != nil {
		log.Printf("Failed to fetch latest binary: %s, attempting to use cockroach binary from your PATH", err)
		cockroachBinary = "cockroach"
	} else {
		log.Printf("Using automatically-downloaded binary: %s", cockroachBinary)
	}

	// Force "/tmp/" so avoid OSX's really long temp directory names
	// which get us over the socket filename length limit.
	baseDir, err := ioutil.TempDir("/tmp", "cockroach-testserver")
	if err != nil {
		return nil, fmt.Errorf("could not create temp directory: %s", err)
	}

	mkDir := func(name string) (string, error) {
		path := filepath.Join(baseDir, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", fmt.Errorf("could not create %s directory: %s: %s", name, path, err)
		}
		return path, nil
	}
	logDir, err := mkDir(logsDirName)
	if err != nil {
		return nil, err
	}
	certsDir, err := mkDir(certsDirName)
	if err != nil {
		return nil, err
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
			{"cert", "create-client", "root"},
		} {
			if err := exec.Command(cockroachBinary, append(args, certArgs...)...).Run(); err != nil {
				return nil, err
			}
		}
		secureOpt = "--certs-dir=" + certsDir
	}

	args := []string{
		cockroachBinary,
		"start",
		"--logtostderr",
		secureOpt,
		"--host=localhost",
		"--port=0",
		"--http-port=0",
		"--store=" + baseDir,
		"--listening-url-file=" + listeningURLFile,
	}

	ts := &testServerImpl{
		serverArgs:       *serverArgs,
		state:            stateNew,
		baseDir:          baseDir,
		cmdArgs:          args,
		stdout:           filepath.Join(logDir, "cockroach.stdout"),
		stderr:           filepath.Join(logDir, "cockroach.stderr"),
		listeningURLFile: listeningURLFile,
		curTenantID:      firstTenantID,
	}
	ts.pgURL.set = make(chan struct{})
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
func (ts *testServerImpl) WaitForInit(db *sql.DB) error {
	var err error
	for i := 0; i < 50; i++ {
		if _, err = db.Query("SHOW DATABASES"); err == nil {
			return err
		}
		log.Printf("WaitForInit: %v", err)
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
			return fmt.Errorf("unexpected error while reading listening URL file: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	u, err := url.Parse(string(bytes.TrimSpace(data)))
	if err != nil {
		return fmt.Errorf("failed to parse SQL URL: %v", err)
	}
	ts.setPGURL(u)

	return nil
}

// Start runs the process, returning an error on any problems,
// including being unable to start, but not unexpected failure.
// It should only be called once in the lifetime of a TestServer object.
func (ts *testServerImpl) Start() error {
	ts.mu.Lock()
	if ts.state != stateNew {
		ts.mu.Unlock()
		return errors.New("Start() can only be called once")
	}
	ts.state = stateRunning
	ts.mu.Unlock()

	ts.cmd = exec.Command(ts.cmdArgs[0], ts.cmdArgs[1:]...)
	ts.cmd.Env = []string{"COCKROACH_MAX_OFFSET=1ns"}

	if len(ts.stdout) > 0 {
		wr, err := newFileLogWriter(ts.stdout)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %s", ts.stdout, err)
		}
		ts.stdoutBuf = wr
	}
	ts.cmd.Stdout = ts.stdoutBuf

	if len(ts.stderr) > 0 {
		wr, err := newFileLogWriter(ts.stderr)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %s", ts.stderr, err)
		}
		ts.stderrBuf = wr
	}
	ts.cmd.Stderr = ts.stderrBuf

	for k, v := range defaultEnv() {
		ts.cmd.Env = append(ts.cmd.Env, k+"="+v)
	}

	err := ts.cmd.Start()
	if ts.cmd.Process != nil {
		log.Printf("process %d started: %s", ts.cmd.Process.Pid, strings.Join(ts.cmdArgs, " "))
	}
	if err != nil {
		log.Print(err.Error())
		if err := ts.stdoutBuf.Close(); err != nil {
			log.Printf("failed to close stdout: %s", err)
		}
		if err := ts.stderrBuf.Close(); err != nil {
			log.Printf("failed to close stderr: %s", err)
		}

		ts.mu.Lock()
		ts.state = stateFailed
		ts.mu.Unlock()

		return fmt.Errorf("failure starting process: %s", err)
	}

	go func() {
		err := ts.cmd.Wait()

		if err := ts.stdoutBuf.Close(); err != nil {
			log.Printf("failed to close stdout: %s", err)
		}
		if err := ts.stderrBuf.Close(); err != nil {
			log.Printf("failed to close stderr: %s", err)
		}

		ps := ts.cmd.ProcessState
		sy := ps.Sys().(syscall.WaitStatus)

		log.Printf("Process %d exited with status %d: %v", ps.Pid(), sy.ExitStatus(), err)
		log.Print(ps.String())

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
				log.Printf("%v", err)
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
		log.Fatal("Stop() called, but Start() was never called")
	}
	if ts.state == stateFailed {
		log.Fatalf("Stop() called, but process exited unexpectedly. Stdout:\n%s\nStderr:\n%s\n",
			ts.Stdout(), ts.Stderr())
		return
	}

	if ts.state != stateStopped {
		// Only call kill if not running. It could have exited properly.
		_ = ts.cmd.Process.Kill()
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
