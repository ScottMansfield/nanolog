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

// MaxLoggers is the maximum number of different loggers that are allowed
const MaxLoggers = 10240

var (
	loggers       = make([]logger, MaxLoggers)
	curLoggersIdx = new(uint32)
)

type logger struct {
	// track varargs lengths and types that are needed
}

var out io.Writer

// Init will set up efficient writing for the log to the output stream given.
// A raw IO stream is best.
func Init(w io.Writer) {
	out = bufio.NewWriter(w)
}

// AddLogger initializes a logger and returns a handle for future logging
func AddLogger(fmt string) uint32 {
	// save some kind of string format to the file
	idx := atomic.AddUint32(curLoggersIdx, 1) - 1

	loggers[idx] = logger{}

	return idx
}
