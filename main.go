package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"log"
	"path"
	"strings"
	"text/template"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	ok := true
	for _, f := range flag.Args() {
		state := newState(f)
		outFname := NoExt(f)
		if outFname == f {
			log.Println("skipping", f, ": input and output file are the same")
			continue
		}
		output := []byte(state.Inc(f))
		state.check(ioutil.WriteFile(outFname, output, 0666))
		if state.err != nil {
			log.Println(f, ":", state.err)
			ok = false
		} else {
			log.Println(outFname, "OK")
		}
	}

	if !ok {
		log.Fatal("exiting with errors")
	}
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

// Retrieves an argument passed to Inc.
func (s *State) Arg(index int) interface{} {
	if index < len(s.args) {
		return s.args[index]
	} else {
		//log.Println("warning: in", s.File, "argument", index, "not provided")
		return ""
	}
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

// Concatenate arguments into single string
func (s *State) Cat(x ...interface{}) string {
	return fmt.Sprint(x...)
}

// Convert argument to lower case.
func (s *State) ToLower(x interface{}) string {
	return strings.ToLower(fmt.Sprint(x))
}

// Cmd executes an external program with arguments.
func(s*State)Cmd(cmd string, args...string)string{
	c := exec.Command(cmd, args...)
	log.Println(cmd, args)
	out, err := c.CombinedOutput()
	if err != nil{
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
