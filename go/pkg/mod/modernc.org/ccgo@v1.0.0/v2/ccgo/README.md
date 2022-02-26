# Table of Contents

1. Usage
1. Installation
1. Changelog

# ccgo

Command ccgo is a C compiler targeting Go.

### Usage

    $ ccgo [options] [files]
    
    -c                          Compile and assemble, but do not link
    -dM                         With -E: generate a list of ‘#define’ directives
                                for all the macros defined during the execution
                                of the preprocessor, including predefined macros.
    -D<macro>[=<val>]           Define a <macro> with <val> as its value.  If
                                just <macro> is given, <val> is taken to be 1
    -e ADDRESS, --entry ADDRESS Set start address (ignored)
    -E                          Preprocess only; do not compile, assemble or link
    -ffreestanding              Do not assume that standard C libraries and
                                "main" exist
    -fomit-frame-pointer        When possible do not generate stack frames (ignored)
    -fPIC                       Generate position-independent code if possible
    --help                      Display this information
    -g --gen-debug              generate debugging information (ignored)
    -h FILENAME, -soname FILENAME
                                Set internal name of shared library
    -I <dir>                    Add <dir> to the end of the main include path
    -l LIBNAME, --library LIBNAME
                                Search for library LIBNAME
    -L DIRECTORY, --library-path DIRECTORY
                                Add DIRECTORY to library search path
    -m64                        Generate 64bit x86-64 code
    -nostdlib                   Do not look for object files in standard path (ignored)
    -o <file>                   Place the output into <file>. Use .go extension
                                to produce a Go source file instead of a binary.
    -O                          Optimize output file (ignored)
    -rpath PATH                 Set runtime shared library search path
    -shared                     Create a shared library
    -v                          Display the programs invoked by the compiler
    --version                   Display compiler version information
    --warn-go-build             Report 'go build' errors as warning
    --warn-unresolved-libs      Report unresolved libraries as warnings
    --warn-unresolved-symbols   Report unresolved symbols as warnings
    -W  --no-warn               suppress warnings (ignored)
    -Wall                       Enable most warning messages (ignored)
    -Wl,<options>               Pass comma-separated <options> on to the linker
    -x <language>               Specify the language of the following input files.
                                Permissible languages include: c.
    
    --ccgo-define-values        Emit #defines that evaluate to a constant
    --ccgo-full-paths           Keep full source code positions instead of
                                basenames
    --ccgo-go                   Do not remove the Go source file used to link the
                                executable file and print its path
    --ccgo-import <paths>       Add import comma separated paths
    --ccgo-pkg-name             Set output Go file package name
    --ccgo-struct-checks        Generate code to verify struct/union sizes
                                and field offsets.
    --ccgo-use-import <exprs>   Add import usage comma separated expressions
    --ccgo-watch                Enable run time watch instrumentation

### Installation

To install or update ccgo and its accompanying tools

     $ go get [-u] modernc.org/ccgo/v2/...

Online documentation: [godoc.org/modernc.org/ccgo/v2/ccgo](http://godoc.org/modernc.org/ccgo/v2/ccgo)

### Changelog

TODO
