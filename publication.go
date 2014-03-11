package main

// Plug-in that parses citation files (.ciw, RIS format)
// and generates a nice publication list.

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
)

func (s *State) Publist(dir string) string {
	pubs, err := LoadPublications(dir)
	if err != nil {
		panic(err)
	}
	out := ""
	for _, p := range pubs {
		out += s.Inc("publication", p.Title, p.Author, p.Journal, p.Date+" "+p.Year, p.DOI, p.Abstract)
	}
	return out
}

// stores the content of a parsed RIS file.
type pub struct {
	Author     []string
	Title      string
	Abstract   string
	Journal    string
	Date, Year string
	RIS        string
	DOI        string
}

// Load publication .ciw files from the directory and serve them under that directory name.
// To be called before LoadContent. TODO: should be OK to call after loadcontent.
// To be called after PubXRefAuthor calls, if any.
func LoadPublications(dir string) ([]*pub, error) {
	d, err1 := os.Open(dir)
	if err1 != nil {
		return nil, err1
	}
	ls, err2 := d.Readdir(-1)
	if err2 != nil {
		return nil, err2
	}

	var pubs []*pub
	var err error

	for _, f := range ls {
		if path.Ext(f.Name()) == ".ciw" {
			fullname := dir + "/" + f.Name()
			pub, e := parseRIS(fullname)
			if e != nil {
				log.Println("error parsing", fullname, ":", e)
				err = e
				continue
			}
			pubs = append(pubs, pub)
		}
	}

	sort.Sort(publist(pubs))
	return pubs, err
}

// makes publication list sortable
type publist []*pub

func (p publist) Len() int           { return len(p) }
func (p publist) Less(i, j int) bool { return p[i].Year > p[j].Year } // most recent first
func (p publist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func parseRIS(fname string) (p *pub, err error) {
	defer func() {
		if e := recover(); e != nil {
			p = nil
			err = fmt.Errorf("%v", e)
		}
	}()
	p = new(pub)
	p.RIS = fname
	f, e := os.Open(fname)
	if e != nil {
		return nil, e
	}
	in := bufio.NewReader(f)

	l, _, e2 := in.ReadLine()
	if e2 != nil {
		return nil, e2
	}
	key, val := string(l[:2]), string(l[3:])
	for len(l) > 3 {
		p.Add(key, val)

		l, _, err = in.ReadLine()
		k := string(l[:2])
		if k != "  " { // keep previous key if empty
			key = k
		}
		if len(l) > 2 {
			val = string(l[3:])
		}
	}
	if err != nil {
		p = nil
	}
	return p, err
}

// add a key/value pair (one line of the RIS file) to the data already stored in pub
func (p *pub) Add(key, val string) {
	switch key {
	case "AF":
		p.Author = append(p.Author, val)
	case "TI":
		p.Title = p.Title + " " + val
	case "AB":
		p.Abstract = p.Abstract + " " + val
	case "JI":
		p.Journal = val
	case "PD":
		p.Date = val
	case "PY":
		p.Year = val
	case "DI":
		p.DOI = "http://doi.org/" + val
	}
}
