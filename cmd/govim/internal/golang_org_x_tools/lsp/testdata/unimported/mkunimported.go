// +build ignore

// mkunimported generates the unimported.go file, containing the Go standard
// library packages.
// The completion items from the std library are computed in the same way as in the
// github.com/govim/govim/cmd/govim/internal/golang_org_x_tools/imports.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	pkgpath "path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

func mustOpen(name string) io.Reader {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func api(base string) string {
	return filepath.Join(runtime.GOROOT(), "api", base)
}

var sym = regexp.MustCompile(`^pkg (\S+).*?, (?:var|func|type|const) ([A-Z]\w*)`)

var unsafeSyms = map[string]bool{"Alignof": true, "ArbitraryType": true, "Offsetof": true, "Pointer": true, "Sizeof": true}

func main() {
	var buf bytes.Buffer
	outf := func(format string, args ...interface{}) {
		fmt.Fprintf(&buf, format, args...)
	}
	outf("// Code generated by mkstdlib.go. DO NOT EDIT.\n\n")
	outf("package unimported\n")
	outf("func _() {\n")
	f := io.MultiReader(
		mustOpen(api("go1.txt")),
		mustOpen(api("go1.1.txt")),
		mustOpen(api("go1.2.txt")),
		mustOpen(api("go1.3.txt")),
		mustOpen(api("go1.4.txt")),
		mustOpen(api("go1.5.txt")),
		mustOpen(api("go1.6.txt")),
		mustOpen(api("go1.7.txt")),
		mustOpen(api("go1.8.txt")),
		mustOpen(api("go1.9.txt")),
		mustOpen(api("go1.10.txt")),
		mustOpen(api("go1.11.txt")),
		mustOpen(api("go1.12.txt")),
	)
	sc := bufio.NewScanner(f)

	pkgs := map[string]bool{
		"unsafe":     true,
		"syscall/js": true,
	}
	paths := []string{"unsafe", "syscall/js"}
	for sc.Scan() {
		l := sc.Text()
		has := func(v string) bool { return strings.Contains(l, v) }
		if has("struct, ") || has("interface, ") || has(", method (") {
			continue
		}
		if m := sym.FindStringSubmatch(l); m != nil {
			path, _ := m[1], m[2]

			if _, ok := pkgs[path]; !ok {
				pkgs[path] = true
				paths = append(paths, path)
			}
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
	sort.Strings(paths)

	var markers []string
	for _, path := range paths {
		marker := strings.ReplaceAll(path, "/", "slash")
		markers = append(markers, marker)
	}
	outf("	//@complete(\"\", %s)\n", strings.Join(markers, ", "))
	outf("}\n")
	outf("// Create markers for unimported std lib packages. Only for use by this test.\n")

	for i, path := range paths {
		name := pkgpath.Base(path)
		marker := markers[i]
		outf("/* %s *///@item(%s, \"%s\", \"\\\"%s\\\"\", \"package\")\n", name, marker, name, path)
	}

	fmtbuf, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("unimported.go", fmtbuf, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
