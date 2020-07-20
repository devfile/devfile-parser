package main

import (
	"fmt"

	"github.com/redhat-developer/devfile-parser/pkg/devfile/parser"
	devfileParser "github.com/redhat-developer/devfile-parser/pkg/devfile/parser"
)

type DevfileObject struct {
	devfileObj parser.DevfileObj
}

func main() {
	devfile, err := ParseDevfile("devfile.yaml")
	if err != nil {
		fmt.Println(err)
	} else {
		for _, component := range devfile.Data.GetAliasedComponents() {
			if component.Dockerfile != nil {
				fmt.Println(component.Dockerfile.Destination)
			}
		}
	}

}

func ParseDevfile(devfileLocation string) (devfileoj parser.DevfileObj, err error) {

	var devfile devfileParser.DevfileObj
	devfile, err := devfileParser.ParseAndValidate(devfileLocation)
	return devfile, err
}
