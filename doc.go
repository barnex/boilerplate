/*
 Boilerplate renders Go templates aimed at static html generation.
 See http://golang.org/pkg/text/template/ for Go's template syntax. 

 	boilerplate *.html.t

 renders all *.html.t template files to the corresponding *.html file.
 The template is fed a *State variable providing template methods.
 The most common one is:

 	{{.Inc "othertemplate"}}

 which recursively renders othertemplate and all templates therein.
*/
package main
