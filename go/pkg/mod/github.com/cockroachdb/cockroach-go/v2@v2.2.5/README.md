# CockroachDB Go Helpers  [![Go Reference](https://pkg.go.dev/badge/github.com/cockroachdb/cockroach-go/v2/crdb.svg)](https://pkg.go.dev/github.com/cockroachdb/cockroach-go/v2/crdb)

This project contains helpers for CockroachDB users writing in Go:
- `crdb` and its subpackages provide wrapper functions for retrying transactions that fail
  due to serialization errors. It is intended for use within any Go application. See
  `crdb/README.md` for more details.
- `testserver` provides functions for starting and connecting to a locally running instance of
  CockroachDB. It is intended for use in test code.

## Prerequisites

The current release (v2) of this library requires Go modules.

You can import it in your code using:

```
import (
	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
)
```
