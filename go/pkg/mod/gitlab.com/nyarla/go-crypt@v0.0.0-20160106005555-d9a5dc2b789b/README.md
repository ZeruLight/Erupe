crypt
=====

A golang implementation of crypt(3).


[![Build Status](https://travis-ci.org/nyarla/go-crypt.svg?branch=master)](https://travis-ci.org/nyarla/go-crypt) [![GoDoc](http://godoc.org/github.com/nyarla/go-crypt?status.svg)](https://godoc.org/github.com/nyarla/go-crypt)

EXAMPLES CODE
-------------

```go
import (
    "fmt"

    "github.com/nyarlabo/go-crypt"
)

func main() {
    fmt.Println(crypt.Crypt("testtest", "es")); // esDRYJnY4VaGM
}
```

WHY I FROKED IT?
----------------

Original implementation is writte by iasija at 2009-12-08,
and original implementation is not supported golang 1.1 or later.

So I fork it for fix this issue, and I added documenation and test code.

Original implementation is hosting on [code.google.com/p/go-crypt](https://code.google.com/p/go-crypt),
and that source code is under the 3-Clause BSD.

NOTE: I could't find to iasija's contact address.

COPYRIGTS AND LICENSE
---------------------

  1. Original Implementation: Copyright (c) 2009 iasija All Rights Reserved. ([BSD-3-Clause](http://opensource.org/licenses/BSD-3-Clause))
  2. Modification Codes: Copyright (c) 22013-2015 Naoki OKAMURA a.k.a nyarla <nyarla@thotep.net> Some Rights Reserved. ([BSD-3-Clause](http://opensource.org/licenses/BSD-3-Clause))

