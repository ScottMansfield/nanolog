[![GoDoc](https://godoc.org/github.com/ScottMansfield/nanolog?status.svg)](https://godoc.org/github.com/ScottMansfield/nanolog)
[![GoReportCard](https://goreportcard.com/badge/github.com/ScottMansfield/nanolog)](https://goreportcard.com/report/github.com/ScottMansfield/nanolog)
[![Coverage Status](https://coveralls.io/repos/github/ScottMansfield/nanolog/badge.svg?branch=master)](https://coveralls.io/github/ScottMansfield/nanolog?branch=master)

# nanolog

"Nanosecond scale" logger inspired by https://github.com/PlatformLab/NanoLog

## Why?

It's about 3.75x faster than the equivalent stdlib `log` package usage and the output log files are about 1/2 the size. These ratios should increase and decrease, respectively, as the amount of unchanging data in each log line increases.

The `AddLogger` method returns a `nanolog.Handle` to an internal data structure that keeps track of the data required to ensure proper operation. This handle is just a simple number identifier. You can think of `AddLogger` like adding a prepared statement in a database. You supply the unchanging information up front, and the system holds on to that while you give it to the changing data. Overall this is much more efficient because less data is transferred.

## Usage

### Logging at runtime

Add loggers by registering them in an init function in any package using `AddLogger`. The main package should set the writer for the logging system (using the `SetWriter` method) before doing much of anything else, as log writes are buffered in memory until the writer is set. Writes include the data `AddLogger` generates, so by the time `main` gets started there's likely data waiting. Log lines are written using the `Log` method.

At the end of the `main` method in your program, you should ensure that you call `nanolog.Flush()` to ensure that the data that has been logged is sent to the writer you supplied. Otherwise, some data may get lost.

```go
package main

import (
	"os"
	"github.com/ScottMansfield/nanolog"
)

var h nanolog.Handle

func init() {
	nanolog.SetWriter(os.Stderr)
	h = nanolog.AddLogger("Example %i32 log %{s} line %c128")
}

func main() {
	nanolog.Log(h, int32(4), "this is a string", 4i)
	nanolog.Flush()
}

```

### Inflating the logs

The logs are written in an efficient format and are thus not human-readable. In order to be able to read them, you will need to "inflate" them. Each log file is self-contained, so the tooling doesn't need any external information to parse the file.

First, compile the `inflate` tool, then use it on the log output file. The tool outputs to stdout, so if you want to save the output for later, make sure to direct it to another file. The following example assumes your log output is stored in `foo.clog`.

```
$ go build github.com/ScottMansfield/nanolog/cmd/inflate
$ ./inflate -f foo.clog > foo-inflated.log
```

## Format

The logger is created with a string format. The interpolation tokens are prefixed using a percentage sign (`%`) and surrounded by optional curly braces when you need to disambiguate. This can be useful if you want to interpolate an `int` but for some reason need to put a number after it that might confuse the system, like a 1, 3, or 6.

```
nanolog.AddLogger("Disambiguate this: %{i}32")
```

In order to output a literal `%`, you use two of them in a row to escape the second one.

## Types

The types that can be interpolated are limited, for now, to those in the following table. The corresponding interpolation tokens are listed next
to each type.

| Type       | Token |
|------------|-------|
| Bool       | b     |
| Int        | i     |
| Int8       | i8    |
| Int16      | i16   |
| Int32      | i32   |
| Int64      | i64   |
| Uint       | u     |
| Uint8      | u8    |
| Uint16     | u16   |
| Uint32     | u32   |
| Uint64     | u64   |
| Float32    | f32   |
| Float64    | f64   |
| Complex64  | c64   |
| Complex128 | c128  |
| String     | s     |

The logging system is strict when it comes to types. For example, an `int16` will not work in a slot meant for an `int`.

## Benchmark

This benchmark is in the `nanolog_test.go` file. It compares the following log line time to log for both `nanolog` and the stdlib `log` package.

nanolog:
```
foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.
```

stdlib:
```
foo thing bar thing %d. Fubar %s foo. sadfasdf %d sdfasfasdfasdffds %d.
```

```
$ go test -bench CompareToStdlib -count 100 >> bench
$ benchstat bench
name                       time/op
CompareToStdlib/Nanolog-8  120ns ± 3%
CompareToStdlib/Stdlib-8   452ns ± 3%
```