/*
 Boilerplate renders Go templates aimed at static html generation.

 	boilerplate *.html.t

 renders all *.html.t template files to the corresponding *.html file.
 The template is fed a *State variable providing template methods.
 The most common one is:

 	{{.Render "othertemplate.h"}}

 which recursively renders othertemplate.h and all templates therein.
*/
package main
