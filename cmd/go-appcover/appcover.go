package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("missing sub-command argument")
		return
	}
	subcmd := os.Args[1]
	runBinary := true
	switch subcmd {
	case "run":
		// Build and run.
	case "build":
		runBinary = false
	default:
		log.Printf("unknown sub-command: %v", subcmd)
		return
	}
	// Remove sub-command from the os.Args.
	os.Args = append(os.Args[:1], os.Args[2:]...)

	f, err := os.Create("appcover_main_test.go")
	if err != nil {
		log.Printf("open appcover test file: %v", err)
		return
	}
	defer func() {
		f.Close()
		if err := os.Remove(f.Name()); err != nil {
			fmt.Printf("cleanup: %v", err)
		}
	}()

	coverageFilename := "_appcover.out"
	if err := generateAppcoverTest(f, coverageFilename); err != nil {
		log.Printf("generate appcover test file: %v", err)
		return
	}

	// Build the test binary.
	log.Println("building test binary...")
	tmp := os.TempDir()
	if tmp == "" || tmp == "/" {
		log.Printf("suspicious temporary dir: %q", tmp)
		return
	}
	binaryFilename := filepath.Join(tmp, "_appcover")
	args := []string{
		"test",
		"-o", binaryFilename,
		"-c",
		"-tags", "appcover",
	}
	args = append(args, os.Args[1:]...)
	args = append(args, ".")
	out, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		log.Printf("build test binary: %v: %s", err, out)
		return
	}

	if runBinary {
		log.Println("running test binary...")
		if err := runApp(binaryFilename); err != nil {
			log.Printf("run test binary: %v", err)
			return
		}
	}
}

func generateAppcoverTest(f *os.File, coverageFilename string) error {
	data := map[string]interface{}{
		"CoverageFilename": coverageFilename,
	}

	tmpl := `//+ build appcover

package main

import "testing"
import "os/signal"
import "flag"
import "io"
import "os"
import "time"

func TestMain(m *testing.M) {
	// Override default -coverprofile flag value.
	flag.Set("test.coverprofile", "_appcover1.out")
	nextID := 2

	// Because main could do os.Exit or it can be killed
	// in any other way unexpectedly, flush
	// coverage data once in a while.
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			flushReport()
			if nextID == 1 {
				flag.Set("test.coverprofile", "_appcover1.out")
				nextID = 2
			} else {
				flag.Set("test.coverprofile", "_appcover2.out")
				nextID = 1
			}
		}
	}()

	go main()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	flushReport()
}

func flushReport() {
	testing.MainStart(&testDeps{}, nil, nil, nil).Run()
}

type testDeps struct{}

func (d testDeps) MatchString(pat, str string) (bool, error)   { return false, nil }
func (d testDeps) StartCPUProfile(io.Writer) error             { return nil }
func (d testDeps) StopCPUProfile()                             {}
func (d testDeps) WriteHeapProfile(io.Writer) error            { return nil }
func (d testDeps) WriteProfileTo(string, io.Writer, int) error { return nil }
func (d testDeps) ImportPath() string                          { return "" }
func (d testDeps) StartTestLog(io.Writer)                      {}
func (d testDeps) StopTestLog() error                          { return nil }
`
	return template.Must(template.New("appcover_main_test").Parse(tmpl)).Execute(f, data)
}

func stat(name string) os.FileInfo {
	info, err := os.Stat(name)
	if err != nil {
		return nil
	}
	return info
}

func chooseProfile(p1, p2 os.FileInfo) os.FileInfo {
	if p1 == nil || p2 == nil {
		return p2
	}

	// Both are non-nil.

	if p1.Size() == 0 {
		return p2
	}
	if p2.Size() == 0 {
		return p1
	}

	if p1.ModTime().Unix() < p2.ModTime().Unix() {
		return p2
	}
	return p1

}

func runApp(app string) error {
	out, err := exec.Command(app).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	p1 := stat("_appcover1.out")
	p2 := stat("_appcover2.out")
	p := chooseProfile(p1, p2)
	if p == nil {
		return fmt.Errorf("can't find non-empty coverage profiles")
	}
	if err := os.Rename(p.Name(), "_appcover.out"); err != nil {
		return err
	}
	if p1 == p {
		removeFile(p2.Name())
	} else {
		removeFile(p1.Name())
	}
	return nil
}

func removeFile(name string) {
	if err := os.Remove(name); err != nil {
		log.Printf("error: %v", err)
	}
}
