package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() == 0 {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("compiling *.t in all subdirectories of", wd)
		recursive(wd)
	} else {
		handleargs(flag.Args())
	}

}

func handleargs(args []string) {
	ok := true
	for _, f := range flag.Args() {
		err := handle(f)
		if err != nil {
			ok = false
		}
	}
	if !ok {
		log.Fatal("exiting with errors")
	}
}

func recursive(dir string) {
	ok := true
	filepath.Walk(dir, func(path string, i os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".t") {
			err := handle(path)
			if err != nil {
				ok = false
			}
		}
		return nil
	})
	if !ok {
		log.Fatal("exiting with errors")
	}
}

func handle(f string) error {
	fmt.Print(f, ": ")
	state := newState(f)
	outFname := NoExt(f)
	if outFname == f {
		return fmt.Errorf("skipping", f, ": input and output file are the same")
	}
	output := []byte(state.Inc(f))
	state.check(ioutil.WriteFile(outFname, output, 0666))
	if state.err != nil {
		fmt.Println(state.err)
	} else {
		fmt.Println("OK")
	}
	return state.err
}

func newState(rootFile string) *State {
	return &State{File: rootFile, vars: make(map[string]interface{})}
}

// a *State is passed to template rendering
type State struct {
	File   string
	parent *State
	args   []interface{}
	vars   map[string]interface{}
	err    error
}

// Recursively expands the template in file "fname"
// and passes optional arguments which are accessible as
// 	{{.Arg 0}} {{.Arg 1}} ...
func (s *State) Inc(fname string, args ...interface{}) string {
	if strings.HasPrefix(fname, "./") {
		fname = s.Dir() + fname
		fname = path.Clean(fname)
		if fname == "." {
			fname = ""
		}
	}
	content := s.Raw(fname)
	t := template.Must(template.New(fname).Parse(content))
	out := bytes.NewBuffer(nil)

	child := newState(s.File)
	child.parent = s
	child.args = args

	s.check(t.Execute(out, child))

	if child.err != nil && s.err == nil {
		s.err = child.err
	}

	return out.String()
}

// Esc is like Raw, but escapes HTML characters.
func (s *State) Esc(fname string) string {
	return template.HTMLEscapeString(s.Raw(fname))
}

// Expands to the raw contents of file "fname", not treating it as a template.
func (s *State) Raw(fname string) string {
	bytes, err := ioutil.ReadFile(fname)
	s.check(err)
	return string(bytes)
}

// Retrieves an argument passed to Inc.
func (s *State) Arg(index int) interface{} {
	if index < len(s.args) {
		return s.args[index]
	} else {
		//log.Println("warning: in", s.File, "argument", index, "not provided")
		return ""
	}
}

func (s *State) Args() []interface{} {
	return s.args
}

// Define a new variable with given name and value.
func (s *State) Def(key string, value interface{}) string {
	s.vars[key] = value
	return "" // must return something to template
}

// Retrieve variable value.
func (state *State) Var(key string) interface{} {
	for s := state; s != nil; s = s.parent {
		if v, ok := s.vars[key]; ok {
			return v
		}
	}
	log.Println("warning:", state.File, "-> undefined variable:", key)
	return ""
}

// Concatenate arguments into single string
func (s *State) Cat(x ...interface{}) string {
	return fmt.Sprint(x...)
}

// Convert argument to lower case.
func (s *State) ToLower(x interface{}) string {
	return strings.ToLower(fmt.Sprint(x))
}

// List all files matching patterns. No patterns returns all files.
// Match in current directory.
func (s *State) Ls(pattern string) []string {
	dir, err := os.Open(s.Dir())
	if err != nil {
		panic(err)
	}
	fi, err2 := dir.Readdir(-1)
	if err2 != nil {
		panic(err2)
	}
	var files []string
	for _, f := range fi {
		match, err := regexp.MatchString(pattern, f.Name())
		if err != nil {
			panic(err)
		}
		if match {
			files = append(files, s.Dir()+f.Name())
		}
	}
	sort.Strings(files)
	return files
}

func (s *State) BaseName(file string) string {
	return path.Base(file)
}

func (s *State) NoExt(file string) string {
	return file[:len(file)-len(path.Ext(file))]
}

// Cmd executes an external program with arguments.
func (s *State) Cmd(cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	log.Println(cmd, args)
	out, err := c.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		log.Println(err)
		return ""
	}
	return string(out)
}

func (s *State) Path() []string {
	d := s.Dir()
	d = d[:len(d)-1] // rm trailing /
	p := strings.Split(d, "/")
	if p[0] == "." {
		p = []string{}
	}
	log.Println("path of", s.File, p)
	return p
}

//func(s*State) BasePath() string{
//	base := ""
//	for _, chr := range s.File{
//		if chr == '/'{
//			base += "../"
//		}
//	}
//	return base
//}

// Dir returns the directory of the currently rendered template file.
func (s *State) Dir() string {
	return path.Dir(s.File) + "/"
}

// Remove extension from file name.
func NoExt(file string) string {
	ext := path.Ext(file)
	return file[:len(file)-len(ext)]
}

func (s *State) check(err error) {
	if err != nil {
		//log.Println("error in", s.File, ": ", err)
		s.err = err
	}
}
