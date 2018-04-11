/*
 * MIT LICENSE
 *
 * Copyright © 2018, G.Ralph Kuntz, MD.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

const (
	primaryTemplate = `
// Copyright (c) 2018 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
// Use of this document is governed by a license found in the LICENSE document.

package react

// {{ .Elem }} is the React element definition corresponding to the HTML <{{ .Name }}> element.
type {{ .Elem }} struct {
	Element
}

// _{{ .Props }} defines the properties for the <{{ .Name }}> element.
type _{{ .Props }} struct {
	*BasicHTMLElement

	{{ range .Attributes }}{{ .Name }} {{ .Type }} ` + "`js:\"{{ .JS }}\"`" + `
	{{ end }}
}

// A creates a new instance of a <{{ .Name }}> element with the provided props and children.
func {{ .Upper }}(props *{{ .Props }}, children ...Element) *{{ .Elem }} {
	rProps := &_{{ .Props }}{
		BasicHTMLElement: newBasicHTMLElement(),
	}

	if props != nil {
		props.assign(rProps)
	}

	return &{{ .Elem }}{
		Element: createElement("{{ .Name }}", rProps, children...),
	}
}
`
	testTemplate = `
// +build js

package react_test

import (
	"testing"

	"honnef.co/go/js/dom"

	"myitcv.io/react"
	"myitcv.io/react/testutils"
)

func Test{{ .Elem }}(t *testing.T) {
	class := "test"

	x := testutils.Wrapper(react.{{ .Upper }}(&react.{{ .Props }}{ClassName: class}))
	cont := testutils.RenderIntoDocument(x)

	el := testutils.FindRenderedDOMComponentWithClass(cont, class)

	if _, ok := el.(*dom.HTMLAnchorElement); !ok {
		t.Fatal("Failed to find <{{ .Name }}> element")
	}
}
`
)

type (
	tomlDesc struct {
		Override   string     `toml:"override"`
		Attributes []tomlAttr `toml:"attributes"`
	}

	tomlAttr struct {
		Name     string `toml:"name"`
		Override string `toml:"override"`
		Type     string `toml:"type"`
	}

	el struct {
		Elem       string // AElem
		Name       string // a
		Props      string // AProps
		Upper      string // A
		Attributes []attr
	}

	attr struct {
		Name string
		JS   string
		Type string
	}
)

var (
	inputFile       = flag.String("i", "elements.toml", ".toml file containing the elements to generate")
	outputDirectory = flag.String("o", ".", "output directory to write the generated Go files")

	elements map[string]tomlDesc
)

func main() {
	flag.CommandLine.Usage = usage
	flag.Parse()

	if _, err := toml.DecodeFile(*inputFile, &elements); err != nil {
		panic(err)
	}

	primary := template.Must(template.New("primary").Parse(primaryTemplate))
	test := template.Must(template.New("test").Parse(testTemplate))

	for k, v := range elements {
		var upper string
		if v.Override != "" {
			upper = v.Override
		} else {
			upper = strings.ToUpper(string(k[0])) + k[1:]
		}

		var attrs []attr
		for _, a := range v.Attributes {
			js := a.Name
			var name string
			if a.Override == "" {
				name = strings.ToUpper(string(js[0])) + js[1:]
			} else {
				name = a.Override
			}
			var t string
			if a.Type == "" {
				t = "string"
			} else {
				t = a.Type
			}
			attrs = append(attrs, attr{Name: name, JS: js, Type: t})
		}

		e := el{
			Elem:       upper + "Elem",
			Name:       k,
			Props:      upper + "Props",
			Upper:      upper,
			Attributes: attrs,
		}

		executeTemplate(k+"_elem.go", primary, e)
		executeTemplate(k+"_elem_test.go", test, e)
	}
}

func executeTemplate(n string, t *template.Template, e el) {
	b := new(bytes.Buffer)
	if err := t.Execute(b, e); err != nil {
		panic(err)
	}

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		panic(err)
	}

	f, err := os.Create(filepath.Join(*outputDirectory, n))
	if err != nil {
		panic(err)
	}

	if _, err := f.Write(formatted); err != nil {
		panic(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
