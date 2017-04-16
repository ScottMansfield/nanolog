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

package reader

import (
	"bytes"
	"testing"

	"github.com/ScottMansfield/nanolog"
)

func TestReader(t *testing.T) {
	inbuf := &bytes.Buffer{}
	nanolog.SetWriter(inbuf)

	// This should exercise every different type in one log line
	h := nanolog.AddLogger("%b %s %i %i8 %i16 %i32 %i64 %u %u8 %u16 %u32 %u64 %f32 %f64 %c64 %c128")
	nanolog.Log(h,
		true, "",
		int(4), int8(4), int16(4), int32(4), int64(4),
		uint(4), uint8(4), uint16(4), uint32(4), uint64(4),
		float32(4), float64(4),
		complex(float32(4), float32(4)), complex(float64(4), float64(4)),
	)
	nanolog.Flush()

	outbuf := &bytes.Buffer{}
	r := New(inbuf, outbuf)
	if err := r.Inflate(); err != nil {
		t.Fatalf("Got error during inflate: %v", err)
	}
}
