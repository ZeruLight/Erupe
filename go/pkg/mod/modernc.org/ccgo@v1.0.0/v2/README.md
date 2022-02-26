# ccgo

Package ccgo translates c99 ASTs to Go. Work In Progress. API unstable.

### Installation


To install or update ccgo and its accompanying tools

     $ go get [-u] modernc.org/ccgo/v2/...

Documentation: [godoc.org/modernc.org/ccgo/v2](http://godoc.org/modernc.org/ccgo/v2)

Building with `make` requires the following Go packages

* github.com/golang/lint/golint
* github.com/mdempsky/maligned
* github.com/mdempsky/unconvert
* honnef.co/go/tools/cmd/unused
* honnef.co/go/tools/cmd/gosimple
* github.com/client9/misspell/cmd/misspell
