/*
 Boilerplate renders Go templates aimed at static html generation.
 See http://golang.org/pkg/text/template/ for Go's template syntax.

 	boilerplate *.html.t

 renders all *.html.t template files to the corresponding *.html file.
 The template is fed a *State variable providing template methods.
 The most common one is:

 	{{.Inc "othertemplate" "arg1" "arg..."}}

 which recursively renders othertemplate and all templates therein.
 Files passed to in .Inc are resolved with respect to the main working directory.

*/
package main
