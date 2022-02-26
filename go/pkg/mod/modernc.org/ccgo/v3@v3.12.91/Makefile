# Copyright 2019 The GOCC Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all clean editor internalError later nuke todo edit 

grep=--include=*.go
ngrep='TODOOK\|internalError\|testdata'

all:
	LC_ALL=C make all_log 2>&1 | tee log

all_log:
	date
	go version
	uname -a
	./unconvert.sh
	gofmt -l -s -w *.go
	GOOS=linux GOARCH=arm go build -o /dev/null
	GOOS=linux GOARCH=386 go build -o /dev/null
	GOOS=linux GOARCH=amd64 go build -o /dev/null
	GOOS=windows GOARCH=386 go build -o /dev/null
	GOOS=windows GOARCH=amd64 go build -o /dev/null
	go vet 2>&1 | grep -v $(ngrep) || true
	golint 2>&1 | grep -v $(ngrep) || true
	make todo
	misspell *.go
	staticcheck | grep -v 'lexer\.go' || true
	maligned || true
	grep -n 'files.*, ok' log
	@grep -n --color=never 'FAIL\|PASS' log
	@grep -n --color=always 'FAIL' log
	grep -n --color=always 'nil pointer' log
	grep -c 'exit status 1' log || true
	grep -c 'exit status 2' log || true
	LC_ALL=C date 2>&1 | tee -a log

clean:
	go clean
	rm -f *~ *.test *.out

edit:
	touch log
	gvim -p Makefile *.go &

editor:
	gofmt -l -s -w *.go
	GO111MODULE=off go build -v -o $(GOPATH)/bin/ccgo

later:
	@grep -n $(grep) LATER * || true
	@grep -n $(grep) MAYBE * || true

nuke: clean
	go clean -i

todo:
	@grep -nr $(grep) ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* * | grep -v $(ngrep) || true
	@grep -nr $(grep) 'TODO\|panic' * | grep -v $(ngrep) || true
	@grep -nr $(grep) BUG * | grep -v $(ngrep) || true
	@grep -nr $(grep) [^[:alpha:]]println * | grep -v $(ngrep) || true
	@grep -nir $(grep) 'work.*progress' || true
