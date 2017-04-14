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
	"encoding/binary"
	"reflect"
	"testing"
)

import "io/ioutil"

func TestAddLogger(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w = bufio.NewWriter(buf)
		h := AddLogger("")

		t.Log("Handle:", h)

		w.Flush()
		out := buf.Bytes()

		if len(out) != 13 {
			t.Fatalf("Expected serialized length of 13 but got %v.\nOutput: % X", len(out), out)
		}

		if out[0] != byte(etLogLine) {
			t.Fatalf("Expected first byte to be etLogLine but got %v", out[0])
		}

		out = out[1:]

		idbuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(idbuf, uint32(h))

		if !bytes.HasPrefix(out, idbuf) {
			t.Fatalf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
		}

		out = out[4:]

		numSegs := binary.LittleEndian.Uint32(out)

		if numSegs != 1 {
			t.Fatalf("Expected 1 segment but got %v", numSegs)
		}

		out = out[4:]

		// no kinds to check here

		seglen := binary.LittleEndian.Uint32(out)

		if seglen != 0 {
			t.Fatalf("Expected segment length of 0 but got %v", seglen)
		}
	})
	t.Run("bool", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w = bufio.NewWriter(buf)
		h := AddLogger("%b")

		t.Log("Handle:", h)

		w.Flush()
		out := buf.Bytes()

		if len(out) != 18 {
			t.Fatalf("Expected serialized length of 18 but got %v.\nOutput: % X", len(out), out)
		}

		if out[0] != byte(etLogLine) {
			t.Fatalf("Expected first byte to be etLogLine but got %v", out[0])
		}

		out = out[1:]

		idbuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(idbuf, uint32(h))

		if !bytes.HasPrefix(out, idbuf) {
			t.Fatalf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
		}

		out = out[4:]

		numSegs := binary.LittleEndian.Uint32(out)

		if numSegs != 2 {
			t.Fatalf("Expected 2 segments but got %v", numSegs)
		}

		out = out[4:]

		// 1 kind (bool), 2 segments
		k := reflect.Kind(out[0])

		if k != reflect.Bool {
			t.Fatalf("Expected kind of reflect.Bool but got %v", k)
		}

		out = out[1:]

		seglen := binary.LittleEndian.Uint32(out)

		if seglen != 0 {
			t.Fatalf("Expected segment length of 0 but got %v", seglen)
		}

		out = out[4:]

		seglen = binary.LittleEndian.Uint32(out)

		if seglen != 0 {
			t.Fatalf("Expected segment length of 0 but got %v", seglen)
		}
	})
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
	if len(segs) != 5 {
		t.Fatalf("Expected 5 segs but got %v", len(segs))
	}
}

func TestLog(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		t.Run("false", func(t *testing.T) {
			buf := &bytes.Buffer{}
			w = bufio.NewWriter(buf)
			h := AddLogger("%b")
			t.Log("Handle:", h)
			w.Flush()
			buf.Reset()

			Log(h, false)

			w.Flush()
			out := buf.Bytes()

			if len(out) != 6 {
				t.Fatalf("Expected serialized length of 6 but got %v.\nOutput: % X", len(out), out)
			}

			if out[0] != byte(etLogEntry) {
				t.Fatalf("Expected first byte to be etLogEntry but got %v", out[0])
			}

			out = out[1:]

			idbuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(idbuf, uint32(h))

			if !bytes.HasPrefix(out, idbuf) {
				t.Fatalf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
			}

			out = out[4:]

			if out[0] != 0 {
				t.Fatalf("Expected false boolean value to be 0 but got %v", out[4])
			}
		})
		t.Run("true", func(t *testing.T) {
			buf := &bytes.Buffer{}
			w = bufio.NewWriter(buf)
			h := AddLogger("%b")
			t.Log("Handle:", h)
			w.Flush()
			buf.Reset()

			Log(h, true)

			w.Flush()
			out := buf.Bytes()

			if len(out) != 6 {
				t.Fatalf("Expected serialized length of 6 but got %v.\nOutput: % X", len(out), out)
			}

			if out[0] != byte(etLogEntry) {
				t.Fatalf("Expected first byte to be etLogEntry but got %v", out[0])
			}

			out = out[1:]

			idbuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(idbuf, uint32(h))

			if !bytes.HasPrefix(out, idbuf) {
				t.Fatalf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
			}

			out = out[4:]

			if out[0] != 1 {
				t.Fatalf("Expected true boolean value to be 1 but got %v", out[0])
			}
		})
	})
}

var testLogHandleSink Handle

func BenchmarkAddLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testLogHandleSink = AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")

		// to prevent it from overflowing the logger array
		*curLoggersIdx = 0
	}
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

func BenchmarkLogParallel(b *testing.B) {
	w = bufio.NewWriter(ioutil.Discard)
	h := AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Log(h, int64(1), "string", uint32(2), uint32(3))
		}
	})
}

func BenchmarkLogSequential(b *testing.B) {
	w = bufio.NewWriter(ioutil.Discard)
	h := AddLogger("foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32.")

	for i := 0; i < b.N; i++ {
		Log(h, int64(1), "string", uint32(2), uint32(3))
	}
}
