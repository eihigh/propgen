package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"golang.org/x/xerrors"
)

func main() {

	app := cli.NewApp()
	app.Name = "propgen"
	app.Action = func(c *cli.Context) error {

		if c.NArg() > 1 {
			return xerrors.New("too many args")
		}

		if c.NArg() == 0 {
			files, err := ioutil.ReadDir(".")
			if err != nil {
				return err
			}
			for _, file := range files {
				path := file.Name() // filepath.Join(dir, file.Name())
				if err := generateFile(path); err != nil {
					return err
				}
			}
		} else {
			path := c.Args()[0]
			if err := generateFile(path); err != nil {
				return err
			}
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("propgen: %+v\n", err)
		os.Exit(1)
	}
}

func generateFile(path string) error {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return xerrors.Errorf("cannot read file %v", path)
	}

	result, err := generate(src)
	if err != nil {
		return xerrors.Errorf("generate from file %v failed", path)
	}

	if result == "" {
		return nil
	}

	// make output path "src_propgen.go"
	ext := filepath.Ext(path)
	head := strings.TrimSuffix(path, ext)
	outPath := head + "_propgen" + ext

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	out.WriteString(result)
	fmt.Println("successfully generated", outPath)
	return nil
}
