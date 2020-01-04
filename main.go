package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	"github.com/simonswine/crd-jsonnet-gen/pkg/generator"
	"github.com/simonswine/crd-jsonnet-gen/pkg/parser"
)

var log logr.Logger

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "3")
	flag.Parse()
	log = klogr.New().WithName("crd-jsonnet-gen")

	var retCode = 0
	if err := run(os.Args[1:]); err != nil {
		log.Error(err, "failed:")
		retCode = 1
	}

	klog.Flush()
	os.Exit(retCode)
}

type dataSource struct {
	packageToAPIGroup func(string) []string
	types             []*parser.Type
}

func (d *dataSource) Types() []*parser.Type {
	return d.types
}

func (d *dataSource) PackageToAPIGroup(s string) []string {
	return d.packageToAPIGroup(s)
}

func run(paths []string) error {
	if len(paths) == 0 {
		return errors.New("no package paths given as arguments")
	}

	log.V(2).Info("search for CRDs in paths", "paths", paths)

	p := parser.New().WithLogger(log)

	for _, path := range paths {
		if err := p.AddDirRecursive(path); err != nil {
			return err
		}
	}

	types, err := p.ExportResources()
	if err != nil {
		return err
	}

	datasource := &dataSource{
		types:             types,
		packageToAPIGroup: p.PackageToAPIGroup,
	}

	generator.WithLogger(log)
	if err := generator.Generate(datasource); err != nil {
		return fmt.Errorf("error generating jsonnet: %w", err)
	}

	return nil
}
