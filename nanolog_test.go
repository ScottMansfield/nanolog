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
	"io/ioutil"
	"math/rand"
	"reflect"
	"testing"
)

var testLetters = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString() string {
	n := rand.Intn(10) + 10

	var ret []rune

	for i := 0; i < n; i++ {
		ret = append(ret, testLetters[rand.Intn(len(testLetters))])
	}

	return string(ret)
}

func TestAddLogger(t *testing.T) {
	genTest := func(logLine string, expectedKinds []reflect.Kind, expectedSegs []string) func(*testing.T) {
		return func(t *testing.T) {
			buf := &bytes.Buffer{}
			w = bufio.NewWriter(buf)
			h := AddLogger(logLine)

			//t.Log("Handle:", h)

			w.Flush()
			out := buf.Bytes()

			expLen := 1 + 4 + 4 + len(expectedKinds)
			for _, s := range expectedSegs {
				expLen += 4 + len(s)
			}

			if len(out) != expLen {
				t.Fatalf("Expected serialized length of %v but got %v.\nOutput: % X", expLen, len(out), out)
			}

			if out[0] != byte(ETLogLine) {
				t.Fatalf("Expected first byte to be ETLogLine but got %v", out[0])
			}

			out = out[1:]

			idbuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(idbuf, uint32(h))

			if !bytes.HasPrefix(out, idbuf) {
				t.Fatalf("Expected prefix to match the handle ID.\nExpected: % X\nGot: % X", idbuf, out[:4])
			}

			out = out[4:]

			numSegs := binary.LittleEndian.Uint32(out)

			if numSegs != uint32(len(expectedSegs)) {
				t.Fatalf("Expected %v segments but got %v", len(expectedSegs), numSegs)
			}

			out = out[4:]

			// first check the kinds match
			for i := range expectedKinds {
				k := reflect.Kind(out[0])

				if k != expectedKinds[i] {
					t.Fatalf("Expected kind of %v but got %v", expectedKinds, k)
				}

				out = out[1:]
			}

			for i := range expectedSegs {
				exp := expectedSegs[i]

				seglen := binary.LittleEndian.Uint32(out)

				if seglen != uint32(len(exp)) {
					t.Fatalf("Expected segment length of %v but got %v", len(exp), seglen)
				}

				out = out[4:]

				seg := string(out[:seglen])

				if exp != seg {
					t.Fatalf("Expected segment %v but got %v", exp, seg)
				}

				out = out[seglen:]
			}
		}
	}

	empties := []string{"", ""}

	type testdat struct {
		line     string
		expKinds []reflect.Kind
		expSegs  []string
	}
	type testmap map[string]testdat

	tests := testmap{
		"empty": {
			line:     "",
			expKinds: nil,
			expSegs:  []string{""},
		},
	}

	addKind := func(tm testmap, name, symbol string, kind reflect.Kind) {
		tm[name] = testdat{
			line:     "%" + symbol,
			expKinds: []reflect.Kind{kind},
			expSegs:  empties,
		}
		tm[name+"Brackets"] = testdat{
			line:     "%{" + symbol + "}",
			expKinds: []reflect.Kind{kind},
			expSegs:  empties,
		}

		s1, s2 := randString(), randString()

		tm[name+"WithStrings"] = testdat{
			line:     s1 + "%" + symbol + s2,
			expKinds: []reflect.Kind{kind},
			expSegs:  []string{s1, s2},
		}

		s1, s2 = randString(), randString()

		tm[name+"BracketsWithStrings"] = testdat{
			line:     s1 + "%{" + symbol + "}" + s2,
			expKinds: []reflect.Kind{kind},
			expSegs:  []string{s1, s2},
		}
	}

	addKind(tests, "bool", "b", reflect.Bool)
	addKind(tests, "string", "s", reflect.String)
	addKind(tests, "int", "i", reflect.Int)
	addKind(tests, "int8", "i8", reflect.Int8)
	addKind(tests, "int16", "i16", reflect.Int16)
	addKind(tests, "int32", "i32", reflect.Int32)
	addKind(tests, "int64", "i64", reflect.Int64)
	addKind(tests, "uint", "u", reflect.Uint)
	addKind(tests, "uint8", "u8", reflect.Uint8)
	addKind(tests, "uint16", "u16", reflect.Uint16)
	addKind(tests, "uint32", "u32", reflect.Uint32)
	addKind(tests, "uint64", "u64", reflect.Uint64)
	addKind(tests, "float32", "f32", reflect.Float32)
	addKind(tests, "float64", "f64", reflect.Float64)
	addKind(tests, "complex64", "c64", reflect.Complex64)
	addKind(tests, "complex128", "c128", reflect.Complex128)

	for name, dat := range tests {
		t.Run(name, genTest(dat.line, dat.expKinds, dat.expSegs))
	}
}

func TestParseLogLine(t *testing.T) {
	buf := &bytes.Buffer{}
	w = bufio.NewWriter(buf)
	f := "foo thing bar thing %i64. Fubar %s foo. sadfasdf %u32 sdfasfasdfasdffds %u32."
	l, segs := parseLogLine(f)

	// verify logger kinds
	if len(l.Kinds) != 4 {
		t.Fatalf("Expected 4 kinds in logger but got %v", len(l.Kinds))
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

			if out[0] != byte(ETLogEntry) {
				t.Fatalf("Expected first byte to be ETLogEntry but got %v", out[0])
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

			if out[0] != byte(ETLogEntry) {
				t.Fatalf("Expected first byte to be ETLogEntry but got %v", out[0])
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
	testLoggerSink   Logger
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
