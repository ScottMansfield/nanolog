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
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/ScottMansfield/nanolog"
)

// Reader enables reading of the compressed file format
type Reader struct {
	r *bufio.Reader
	w *bufio.Writer
}

// New creates a new Reader with the given reader and writer
func New(r io.Reader, w io.Writer) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

// Inflate will read from the supplied reader and inflate the contents into the
// supplied writer
func (r *Reader) Inflate() error {
	loggers := make(map[uint32]nanolog.Logger)

	for {
		rawType, err := r.r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		recordType := nanolog.EntryType(rawType)

		switch recordType {
		case nanolog.ETLogLine:
			logger := nanolog.Logger{}

			buf := make([]byte, 4)

			// First comes the line ID
			if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
				return err
			}
			id := binary.LittleEndian.Uint32(buf)

			// Then the number of string segments
			if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
				return err
			}
			numsegs := binary.LittleEndian.Uint32(buf)

			// read in the kinds, numsegs - 1 of them
			for i := uint32(0); i < numsegs-1; i++ {
				b, err := r.r.ReadByte()
				if err != nil {
					return err
				}

				k := reflect.Kind(b)
				logger.Kinds = append(logger.Kinds, k)
			}

			// read in the string segments that surround the interpolations
			for i := uint32(0); i < numsegs; i++ {
				if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
					return err
				}

				strlen := binary.LittleEndian.Uint32(buf)
				strbuf := make([]byte, strlen)

				if _, err := io.ReadAtLeast(r.r, strbuf, len(strbuf)); err != nil {
					return err
				}

				logger.Segs = append(logger.Segs, string(strbuf))
			}

			loggers[id] = logger

		case nanolog.ETLogEntry:
			smallbuf := make([]byte, 2)
			buf := make([]byte, 4)
			longbuf := make([]byte, 8)
			var toWrite interface{}

			// First comes the line ID
			if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
				return err
			}
			id := binary.LittleEndian.Uint32(buf)

			logger := loggers[id]

			r.w.WriteString(logger.Segs[0])

			for i := 1; i < len(logger.Segs); i++ {
				// first read whatever kind data is needed
				switch logger.Kinds[i-1] {
				case reflect.Bool:
					v, err := r.r.ReadByte()
					if err != nil {
						return err
					}

					if v == 0 {
						toWrite = "false"
					} else if v == 1 {
						toWrite = "true"
					}

				case reflect.String:
					// First comes the string length
					if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
						return err
					}
					strlen := binary.LittleEndian.Uint32(buf)

					// copy the string from input to output
					io.Copy(r.w, io.LimitReader(r.r, int64(strlen)))

					toWrite = nil

				// ints
				case reflect.Int:
					fallthrough
				case reflect.Int64:
					if _, err := io.ReadAtLeast(r.r, longbuf, len(longbuf)); err != nil {
						return err
					}

					toWrite = int64(binary.LittleEndian.Uint64(longbuf))

				case reflect.Int8:
					b, err := r.r.ReadByte()
					if err != nil {
						return err
					}

					toWrite = int8(b)

				case reflect.Int16:
					if _, err := io.ReadAtLeast(r.r, smallbuf, len(smallbuf)); err != nil {
						return err
					}

					toWrite = int16(binary.LittleEndian.Uint16(smallbuf))

				case reflect.Int32:
					if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
						return err
					}

					toWrite = int32(binary.LittleEndian.Uint32(buf))

				// uints
				case reflect.Uint:
					fallthrough
				case reflect.Uint64:
					if _, err := io.ReadAtLeast(r.r, longbuf, len(longbuf)); err != nil {
						return err
					}

					toWrite = binary.LittleEndian.Uint64(longbuf)

				case reflect.Uint8:
					b, err := r.r.ReadByte()
					if err != nil {
						return err
					}

					toWrite = uint8(b)

				case reflect.Uint16:
					if _, err := io.ReadAtLeast(r.r, smallbuf, len(smallbuf)); err != nil {
						return err
					}

					toWrite = binary.LittleEndian.Uint16(smallbuf)

				case reflect.Uint32:
					if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
						return err
					}

					toWrite = binary.LittleEndian.Uint32(buf)

				// floats
				case reflect.Float32:
					if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
						return err
					}

					toWrite = math.Float32frombits(binary.LittleEndian.Uint32(buf))

				case reflect.Float64:
					if _, err := io.ReadAtLeast(r.r, longbuf, len(longbuf)); err != nil {
						return err
					}

					toWrite = math.Float64frombits(binary.LittleEndian.Uint64(longbuf))

				// complex
				case reflect.Complex64:
					if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
						return err
					}

					real := math.Float32frombits(binary.LittleEndian.Uint32(buf))

					if _, err := io.ReadAtLeast(r.r, buf, len(buf)); err != nil {
						return err
					}

					imag := math.Float32frombits(binary.LittleEndian.Uint32(buf))

					toWrite = complex(real, imag)

				case reflect.Complex128:
					if _, err := io.ReadAtLeast(r.r, longbuf, len(longbuf)); err != nil {
						return err
					}

					real := math.Float64frombits(binary.LittleEndian.Uint64(longbuf))

					if _, err := io.ReadAtLeast(r.r, longbuf, len(longbuf)); err != nil {
						return err
					}

					imag := math.Float64frombits(binary.LittleEndian.Uint64(longbuf))

					toWrite = complex(real, imag)

				default:
					return fmt.Errorf("Invalid Kind in logger: %v", logger.Kinds[i-1])
				}

				if toWrite != nil {
					_, err = fmt.Fprint(r.w, toWrite)
					if err != nil {
						return err
					}
				}

				r.w.WriteString(logger.Segs[i])
			}

			r.w.WriteByte('\n')

		default:
			return errors.New("Bad file format")
		}
	}

	r.w.Flush()

	return nil
}
