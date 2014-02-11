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
		state.check(ioutil.WriteFile(outFname, []byte(state.Render(f)), 0666))
	}
}

func newState(rootFile string) *State{
	return &State{Root: rootFile, File: rootFile, vars: make(map[string]interface{})}
}

// a *State is passed to template rendering
type State struct {
	File, Root string
	Args []interface{}
	vars map[string]interface{}
}

// Render recursively expands the template in fname.
func (s *State) Render(fname string, args ...interface{}) string {
	content := s.Read(fname)
	t := template.Must(template.New(fname).Parse(content))
	out := bytes.NewBuffer(nil)

	prev := *s

	s.File = fname
	s.check(t.Execute(out, s))

	*s = prev

	return out.String()
}

func (s *State) Def(key string, value interface{}) string {
	s.vars[key] = value
	return "" // must return something to template
}

func (s *State) Var(key string) interface{} {
	if v, ok := s.vars[key]; ok {
		return v
	} else {
		log.Println("warning: undefined variable", key, "in", s.File)
		return ""
	}
}

// Read expands to the raw contents of fname without rendering the file.
func (s *State) Read(fname string) string {
	bytes, err := ioutil.ReadFile(fname)
	s.check(err)
	return string(bytes)
}

// Remove extension from file name.
func NoExt(file string) string {
	ext := path.Ext(file)
	return file[:len(file)-len(ext)]
}

func (s *State) check(err error) {
	if err != nil {
		log.Fatal("error in ", s.File, ": ", err)
	}
}
