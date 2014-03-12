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
		recursive(".")
	} else {
		handleargs(flag.Args())
	}

}

func handleargs(args []string) {
	for _, f := range flag.Args() {
		handle(f)
	}
}

func recursive(dir string) {
	filepath.Walk(dir, func(path string, i os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".t") {
			handle(path)
		}
		return nil
	})
}

func handle(f string) {
	//fmt.Print(f, ": ")
	state := newState(f)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(state.File, ":", err)
			os.Exit(-1)
		}
	}()
	outFname := NoExt(f)
	if outFname == f {
		fmt.Println("skipping", f, ": input and output file are the same")
	}
	output := []byte(state.Inc(f))
	state.check(ioutil.WriteFile(outFname, output, 0666))
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
	//err    error
}

type indent int

func (t indent) String() string {
	str := "\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
	return str[:t]
}

var tabs indent

func short(arg interface{}) string {
	str := fmt.Sprint(arg)
	const maxlen = 20
	if len(str) > maxlen {
		return str[:maxlen-3] + "..."
	} else {
		return str
	}
}

// Recursively expands the template in file "fname"
// and passes optional arguments which are accessible as
// 	{{.Arg 0}} {{.Arg 1}} ...
func (s *State) Inc(fname string, args ...interface{}) string {
	fmt.Print(tabs, "Inc ", fname, "(")
	for _, arg := range args {
		fmt.Print(short(arg), " ")
	}
	fmt.Println(")")

	if fname == "" { // inc nothing
		return ""
	}
	tabs++
	defer func() {
		tabs--
	}()
	fname = s.abs(fname)
	content := s.Raw(fname)
	t := template.Must(template.New(fname).Parse(content))
	out := bytes.NewBuffer(nil)

	child := newState(s.File)
	child.parent = s
	child.args = args

	s.check(t.Execute(out, child))

	return out.String()
}

func (s *State) Echo(msg ...interface{}) string {
	fmt.Print(tabs)
	fmt.Println(msg...)
	return ""
}

// resolve file name: ./name is absolute with respect to s.Dir,
// name (without ./) is relative wrt. wd.
func (s *State) abs(fname string) string {
	if strings.HasPrefix(fname, "./") {
		fname = s.Dir() + fname
		fname = path.Clean(fname)
	}
	return fname
}

// Esc is like Raw, but escapes HTML characters.
func (s *State) Esc(fname string) string {
	return template.HTMLEscapeString(s.Raw(fname))
}

// Expands to the raw contents of file "fname", not treating it as a template.
func (s *State) Raw(fname string) string {
	fname = s.abs(fname)
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
	s.check(err)
	fi, err2 := dir.Readdir(-1)
	s.check(err2)
	var files []string
	for _, f := range fi {
		match, err := regexp.MatchString(pattern, f.Name())
		s.check(err)
		if match {
			files = append(files, s.Dir()+f.Name())
		}
	}
	sort.Strings(files)
	return files
}

// Lists all subdirectories.
func (s *State) LsDirs() []string {
	dir, err := os.Open(s.Dir())
	s.check(err)
	fi, err2 := dir.Readdir(-1)
	s.check(err2)
	var files []string
	for _, f := range fi {
		if f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
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

func (s *State) Exists(file string) bool {
	file = s.abs(file)
	_, err := os.Stat(file)
	if err != nil {
		return false
	}
	return true
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
		panic(fmt.Sprint(s.File, ":", err))
	}
}
