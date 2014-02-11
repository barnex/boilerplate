package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	for _, f := range flag.Args() {
		render(f)
	}
}

func render(fname string) {
	state := &State{fname, make(map[string]interface{})}

	outFname := NoExt(fname)
	out, err := os.OpenFile(outFname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	state.check(err)
	defer out.Close()
	ioutil.WriteFile(outFname, []byte(state.Render(fname)), 0666)

}

type State struct {
	File string
	vars map[string]interface{}
}

func (s *State) Render(fname string) string {
	content := s.Read(fname)
	t := template.Must(template.New(fname).Parse(content))
	out := bytes.NewBuffer(nil)
	prevFile := s.File
	s.File = fname
	s.check(t.Execute(out, s))
	s.File = prevFile
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
