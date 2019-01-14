/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DATA-DOG/godog"
)

func main() {
	fmt.Println("Start Test Run")
	locationoffeaturefiles := flag.String("feature", "", "path to feature files")
	tags := flag.String("tags", "", "tags of testcases to be run")
	flag.Parse()
	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:      "progress",
		Paths:       []string{*locationoffeaturefiles},
		Tags:        *tags,
		Concurrency: 1,
		NoColors:    true,
		Output:      os.Stdout,
	})

	fmt.Println("Finish Test Run")
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {

	s.Step(`^this test case is executed, the user should see "([^"]*)"$`, HelloWorld001)

}
