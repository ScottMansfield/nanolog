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
	"flag"
	"os"

	"github.com/ScottMansfield/nanolog/reader"
)

func main() {
	var fileName string
	flag.StringVar(&fileName, "f", "", "Input file name")
	flag.Parse()

	infile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	if err := reader.New(infile, os.Stdout).Inflate(); err != nil {
		panic(err)
	}
}
