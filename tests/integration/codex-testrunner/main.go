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
	//tags := "@HelloWorld001"
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
