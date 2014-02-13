package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"path"
	"text/template"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	for _, f := range flag.Args() {
		state := newState(f)
		outFname := NoExt(f)
		if outFname == f {
			log.Println("skipping", f, ": input and output file are the same")
			continue
		}
		output := []byte(state.Inc(f))
		if state.err == nil {
			state.check(ioutil.WriteFile(outFname, output, 0666))
		}
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

// Render recursively expands the template in fname.
func (s *State) Inc(fname string, args ...interface{}) string {
	content := s.Raw(fname)
	t := template.Must(template.New(fname).Parse(content))
	out := bytes.NewBuffer(nil)

	child := newState(s.File)
	child.parent = s
	child.args = args

	s.check(t.Execute(out, child))

	return out.String()
}

func (s *State) Def(key string, value interface{}) string {
	s.vars[key] = value
	return "" // must return something to template
}

func (state *State) Var(key string) interface{} {
	for s := state; s != nil; s = s.parent {
		if v, ok := s.vars[key]; ok {
			return v
		}
	}
	log.Println("warning:", state.File, "-> undefined variable:", key)
	return ""
}

// Read expands to the raw contents of fname without rendering the file.
func (s *State) Raw(fname string) string {
	bytes, err := ioutil.ReadFile(fname)
	s.check(err)
	return string(bytes)
}

func (s *State) Arg(index int) interface{} {
	return s.args[index]
}

// Remove extension from file name.
func NoExt(file string) string {
	ext := path.Ext(file)
	return file[:len(file)-len(ext)]
}

func (s *State) check(err error) {
	if err != nil {
		log.Println("error in", s.File, ": ", err)
		s.err = err
	}
}
