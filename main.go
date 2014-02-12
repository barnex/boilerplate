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
		if outFname == f{
			log.Println("skipping", f, ": input and output file are the same")
			continue
		}
		output := []byte(state.Inc(f))
		if state.err == nil{
			state.check(ioutil.WriteFile(outFname, output, 0666))
		}
	}
}

func newState(rootFile string) *State{
	return &State{Root: rootFile, File: rootFile }
}

// a *State is passed to template rendering
type State struct {
	File, Root string
	args []interface{}
	err error
}

// Render recursively expands the template in fname.
func (s *State) Inc(fname string, args ...interface{}) string {
	content := s.Raw(fname)
	t := template.Must(template.New(fname).Parse(content))
	out := bytes.NewBuffer(nil)

	prev := *s

	s.File = fname
	s.args = args
	s.check(t.Execute(out, s))

	*s = prev

	return out.String()
}


// TODO: scoping
//func (s *State) Def(key string, value interface{}) string {
//       s.vars[key] = value
//       return "" // must return something to template
//}
//
//func (s *State) Var(key string) interface{} {
//       if v, ok := s.vars[key]; ok {
//               return v
//       } else {
//               log.Println("warning: undefined variable", key, "in", s.File)
//               return ""
//       }
//}



// Read expands to the raw contents of fname without rendering the file.
func (s *State) Raw(fname string) string {
	bytes, err := ioutil.ReadFile(fname)
	s.check(err)
	return string(bytes)
}

func(s*State)Arg(index int)interface{}{
	return s.args[index]
}

// Remove extension from file name.
func NoExt(file string) string {
	ext := path.Ext(file)
	return file[:len(file)-len(ext)]
}

func (s *State) check(err error) {
	if err != nil {
		log.Println("error in ", s.Root, "->", s.File, ": ", err)
		s.err = err
	}
}
