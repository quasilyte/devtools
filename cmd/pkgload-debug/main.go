package main

import (
	"flag"
	"fmt"
	"go/token"
	"log"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

func main() {
	verbose := flag.Bool("v", false,
		"whether to print more info about the loaded pacakges")
	flag.Parse()
	patterns := flag.Args()

	goModule := getGoModule()
	if goModule != "" {
		fmt.Printf("GO111MODULE=%s\n", goModule)
	}
	fmt.Println(getGoVersion())
	fmt.Println()

	start := time.Now()

	cfg := &packages.Config{
		Mode:  packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes,
		Tests: true,
		Fset:  token.NewFileSet(),
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Panicf("load packages: %v", err)
	}

	if len(patterns) != 0 {
		for _, p := range patterns {
			fmt.Printf("loading %s\n", p)
		}
	} else {
		fmt.Println("running Load() without explicit args")
	}

	fmt.Println()

	if len(pkgs) != 0 {
		fmt.Println("packages:")
		for _, pkg := range pkgs {
			fmt.Printf("* id='%s' name='%s' path='%s' gofiles=%d\n",
				pkg.ID, pkg.Name, pkg.PkgPath, len(pkg.GoFiles))
			if !*verbose {
				continue
			}
			if len(pkg.GoFiles) != 0 {
				fmt.Println("    go files:")
				for _, filename := range pkg.GoFiles {
					fmt.Printf("      %s\n", filename)
				}
			}
		}
	} else {
		fmt.Println("no packages found")
	}

	fmt.Println()
	fmt.Printf("loading took %.2f seconds\n", time.Since(start).Seconds())
}

func getGoVersion() string {
	out, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		log.Panicf("run go version: %v: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func getGoModule() string {
	out, err := exec.Command("go", "env", "GO111MODULE").CombinedOutput()
	if err != nil {
		log.Panicf("run go env: %v: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}
