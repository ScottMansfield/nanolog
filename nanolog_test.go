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

import (
	"bufio"
	"bytes"
	"testing"
)

import "io/ioutil"

func TestAddLogger(t *testing.T) {
	// TODO: actually test
}

func TestParseLogLine(t *testing.T) {
	buf := &bytes.Buffer{}
	w = bufio.NewWriter(buf)
	f := "foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32."
	l, segs := parseLogLine(f)

	// verify logger kinds
	if len(l.kinds) != 4 {
		t.Fatalf("Expected 4 kinds in logger but got %v", len(l.kinds))
	}

	// verify logger segs
	if len(segs) != 4 {
		t.Fatalf("Expected 5 segs but got %v", len(segs))
	}
}

func TestLog(t *testing.T) {
	buf := &bytes.Buffer{}
	w = bufio.NewWriter(buf)
	h := AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")

	//println(len(loggers[h].kinds))
	//for _, k := range loggers[h].kinds {
	//	println(k.String())
	//}

	Log(h, int64(1), "string", uint32(2), uint32(3))
}

func BenchmarkAddLogger(b *testing.B) {
	//
}

var (
	testLoggerSink   logger
	testSegmentsSink []string
)

func BenchmarkParseLogLine(b *testing.B) {
	w = bufio.NewWriter(ioutil.Discard)
	f := "The operation %s could not be completed. Wanted %u64 bar %c128 %b %{s} %{i32}"
	for i := 0; i < b.N; i++ {
		testLoggerSink, testSegmentsSink = parseLogLine(f)
	}
}

func BenchmarkLog(b *testing.B) {
	w = bufio.NewWriter(ioutil.Discard)
	h := AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")
	for i := 0; i < b.N; i++ {
		Log(h, int64(1), "string", uint32(2), uint32(3))
	}
}
