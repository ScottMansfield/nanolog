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

package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ScottMansfield/nanolog"
)

var (
	logWorking  nanolog.Handle
	logFinished nanolog.Handle
)

func init() {
	// Set up nanolog writer
	nanologout, err := os.Create("foo.clog")
	if err != nil {
		panic(err)
	}
	nanolog.SetWriter(nanologout)

	logWorking = nanolog.AddLogger("Worker %u8, working on task %i, attempt %i.")
	logFinished = nanolog.AddLogger("Finished task %i. Result was: %f64, string version %s")
}

func main() {
	wg := &sync.WaitGroup{}

	// Do many goroutines doing lots of work to show difference in data size
	start := time.Now()
	for i := uint8(0); i < 100; i++ {
		wg.Add(1)

		go func(id uint8, wg *sync.WaitGroup) {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < 1000; i++ {
				var v float64
				for j := 0; j < 100; j++ {
					nanolog.Log(logWorking, id, i, j)

					v = r.Float64()
				}

				s := fmt.Sprintf("%f", v)

				nanolog.Log(logFinished, i, v, s)
			}

			wg.Done()
		}(i, wg)
	}

	wg.Wait()

	fmt.Println("Nanolog took", time.Since(start))

	nanolog.Flush()

	// Set up standard logging package writer / value
	start = time.Now()
	regoutraw, err := os.Create("foo.log")
	if err != nil {
		panic(err)
	}
	regout := bufio.NewWriter(regoutraw)
	reg := log.New(regout, "", 0)

	for i := uint8(0); i < 100; i++ {
		wg.Add(1)

		go func(id uint8, wg *sync.WaitGroup) {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < 1000; i++ {
				var v float64
				for j := 0; j < 100; j++ {
					reg.Printf("Worker %d, working on task %d, attempt %d.", id, i, j)

					v = r.Float64()
				}

				s := fmt.Sprintf("%f", v)

				reg.Printf("Finished task %d. Result was: %f, string version %s", i, v, s)
			}

			wg.Done()
		}(i, wg)
	}

	wg.Wait()

	fmt.Println("Stdlib log took", time.Since(start))

	regout.Flush()
}
