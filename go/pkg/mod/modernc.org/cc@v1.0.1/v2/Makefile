# Copyright 2017 The CC Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all clean cover cpu editor internalError later mem nuke todo edit

grep=--include=*.go --include=*.l --include=*.y --include=*.yy
ngrep='TODOOK\|parser\.go\|scanner\.go\|trigraphs\.go\|.*_string\.go'

all: editor
	go vet 2>&1 | grep -v $(ngrep) || true
	golint 2>&1 | grep -v $(ngrep) || true
	make todo
	misspell *.go
	staticcheck || true
	maligned || true
	grep -n 'FAIL\|PASS' log
	go version
	date

clean:
	go clean
	rm -f *~ *.test *.out

cover:
	t=$(shell tempfile) ; go test -coverprofile $$t && go tool cover -html $$t && unlink $$t

cpu: clean
	go test -run @ -bench . -cpuprofile cpu.out
	go tool pprof -lines *.test cpu.out

edit:
	touch log
	gvim -p Makefile *.l *.yy *_test.go log ast2.go cc.go cpp.go encoding.go enum.go etc.go lexer.go model.go operand.go type.go

editor: ast.go parser.go scanner.go trigraphs.go stringer.go
	go version
	./unconvert.sh
	gofmt -l -s -w *.go
	#TODO GOOS=linux GOARCH=arm go build
	#TODO GOOS=linux GOARCH=386 go build
	GOOS=linux GOARCH=amd64 go build
	#TODO GOOS=windows GOARCH=386 go build
	#TODO GOOS=windows GOARCH=amd64 go build
	go test -i
	go test -short 2>&1 | tee log
	go install

internalError:
	egrep -ho '"internal error.*"' *.go | sort | cat -n

later:
	@grep -n $(grep) LATER * || true
	@grep -n $(grep) MAYBE * || true

mem: clean
	go test -run @ -bench . -memprofile mem.out -memprofilerate 1 -timeout 24h
	go tool pprof -lines -web -alloc_space *.test mem.out

nuke: clean
	go clean -i

todo:
	@grep -nr $(grep) ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* * | grep -v $(ngrep) || true
	@grep -nr $(grep) TODO * | grep -v $(ngrep) || true
	@grep -nr $(grep) BUG * | grep -v $(ngrep) || true
	@grep -nr $(grep) [^[:alpha:]]println * | grep -v $(ngrep) || true

ast.go parser.go scanner.go trigraphs.go stringer.go: parser.yy scanner.l trigraphs.l enum.go
	go generate	
