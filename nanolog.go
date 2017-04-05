// Copyright 2017 Scott Mansfield
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nanolog

import "sync/atomic"
import "io"
import "bufio"
import "os"
import "reflect"

// MaxLoggers is the maximum number of different loggers that are allowed
const MaxLoggers = 10240

// The format string is a straightforward format inspired by the full fledged
// fmt.Fprintf function. The codes are unique to thsi package, so normal fmt
// documentation will not be applicable.

// THe format string is similar to fmt in that it uses the percent sign (a.k.a.
// the modulo operator) to signify the start of a format code. The reader is
// greedy, meaning that the parser will attempt to read as much as it can for a
// code before it stops. E.g. if you have a generic int in the middle of your
// format string immediately followed by the number 1 and a space ("%i1 "), the
// parser may complain saying that it encountered an invalid code. To fix this,
// use curly braces after the percent sign to surround the code: "%{i}1 ".

// Kinds from the reflect package and their corresponding format codes
//
// Kind         Code
// ------------------------
// Bool         b
// Int          i
// Int8         i8
// Int16        i16
// Int32        i32
// Int64        i64
// Uint         ui
// Uint8        ui8
// Uint16       ui16
// Uint32       ui32
// Uint64       ui64
// Uintptr
// Float32      f32
// Float64      f64
// Complex64    c64
// Complex128   c128
// Array
// Chan
// Func
// Interface
// Map
// Ptr
// Slice
// String
// Struct
// UnsafePointer

// LogHandle is a simple handle to an internal logging data structure
// LogHandles are returned by the AddLogger method and used by the Log method to
// actually log data.
type LogHandle uint32

var (
	loggers       = make([]logger, MaxLoggers)
	curLoggersIdx = new(uint32)
)

type logger struct {
	// track varargs lengths and types that are needed
	kinds []reflect.Kind
}

var w *bufio.Writer = bufio.NewWriter(os.Stderr)

// SetWriter will set up efficient writing for the log to the output stream given.
// A raw IO stream is best.
func SetWriter(new io.Writer) {
	w.Flush()
	w = bufio.NewWriter(new)
}

// AddLogger initializes a logger and returns a handle for future logging
func AddLogger(fmt string) LogHandle {
	// save some kind of string format to the file
	idx := atomic.AddUint32(curLoggersIdx, 1) - 1

	loggers[idx] = logger{}

	// parse string and build kinds slice

	return LogHandle(idx)
}

// Log logs to the output stream for the logging package
func Log(handle LogHandle, args ...interface{}) error {
	l := loggers[handle]

	if len(l.kinds) != len(args) {
		panic("UH OH")
	}

	for idx := range l.kinds {
		if l.kinds[idx] != reflect.ValueOf(args[idx]).Kind() {
			panic("NO MATCH")
		}

		// write serialized version to writer
	}
}
