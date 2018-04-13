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

	{{ range .Attrs }}{{ .Name }} {{ .Type }} ` + "`js:\"{{ .JS }}\"`" + `
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
	Desc struct {
		Override   string
		Attributes []Attr
	}

	Attr struct {
		Name     string
		Override string
		Type     string
	}

	templElem struct {
		Elem, Name, Props, Upper string
		Attrs                    []templAttr
	}

	templAttr struct {
		Name, JS, Type string
	}
)

var (
	outputDirectory = flag.String("o", ".", "output directory to write the generated Go files")

	// elements contains all of the Go wrappers to generate for the underlying HTML elements.
	// Commented items have already been hand-written.
	elements = map[string]Desc{
		// "a"
		"abbr":    Desc{},
		"acronym": Desc{},
		"address": Desc{},
		"applet": Desc{
			Attributes: []Attr{
				{Name: "align"},
				{Name: "alt"},
				{Name: "archive"},
				{Name: "code"},
				{Name: "codebase"},
				{Name: "datafld", Override: "DataFld"},
				{Name: "datasrc", Override: "DataSrc"},
				{Name: "height"},
				{Name: "hspace", Override: "HSpace"},
				{Name: "mayscript", Override: "MayScript"},
				{Name: "name"},
				{Name: "object"},
				{Name: "src"},
				{Name: "vspace", Override: "VSpace"},
				{Name: "width"},
			},
		},
		"area": Desc{
			Attributes: []Attr{
				{Name: "alt"},
				{Name: "coords"},
				{Name: "download"},
				{Name: "href"},
				{Name: "hreflang", Override: "HrefLang"},
				{Name: "media"},
				{Name: "referrerpolicy", Override: "ReferrerPolicy"},
				{Name: "rel"},
				{Name: "shape"},
				{Name: "target"},
			},
		},
		"article": Desc{},
		"aside":   Desc{},
		"audio": Desc{
			Attributes: []Attr{
				{Name: "autoplay", Override: "AutoPlay"},
				{Name: "buffered"},
				{Name: "controls"},
				{Name: "loop"},
				{Name: "mozCurrentSampleOffset", Override: "MozCurrentSampleOffset"},
				{Name: "muted"},
				{Name: "played"},
				{Name: "preload"},
				{Name: "src"},
				{Name: "volume"},
			},
		},
		"b": Desc{},
		"base": Desc{
			Attributes: []Attr{
				{Name: "href"},
				{Name: "target"},
			},
		},
		"basefont": Desc{
			Attributes: []Attr{
				{Name: "color"},
				{Name: "face"},
				{Name: "size"},
			},
			Override: "BaseFont",
		},
		"bdi": Desc{},
		"bdo": Desc{},
		"blockquote": Desc{
			Attributes: []Attr{
				{Name: "cite"},
			},
			Override: "BlockQuote",
		},
		"body": Desc{
			Attributes: []Attr{
				{Name: "onafterprint", Override: "OnAfterPrint"},
				{Name: "onbeforeprint", Override: "OnBeforePrint"},
				{Name: "onbeforeunload", Override: "OnBeforeUnload"},
				{Name: "onblur", Override: "OnBlur"},
				{Name: "onerror", Override: "OnError"},
				{Name: "onfocus", Override: "OnFocus"},
				{Name: "onhashchange", Override: "OnHashChange"},
				{Name: "onlanguagechange", Override: "OnLanguageChange"},
				{Name: "onload", Override: "OnLoad"},
				{Name: "onmessage", Override: "OnMessage"},
				{Name: "onoffline", Override: "OnOffline"},
				{Name: "ononline", Override: "OnOnline"},
				{Name: "onpopstate", Override: "OnPopState"},
				{Name: "onredo", Override: "OnRedo"},
				{Name: "onresize", Override: "OnResize"},
				{Name: "onstorage", Override: "OnStorage"},
				{Name: "onundo", Override: "OnUndo"},
				{Name: "onunload", Override: "OnUnload"},
			},
		},
		// "br"
		// "button"
		"canvas": Desc{
			Attributes: []Attr{
				{Name: "height"},
				{Name: "width"},
			},
		},
		"caption": Desc{},
		"cite":    Desc{},
		// "code"
		"col": Desc{
			Attributes: []Attr{
				{Name: "bgcolor", Override: "BGColor"},
				{Name: "span"},
			},
		},
		"colgroup": Desc{
			Attributes: []Attr{
				{Name: "bgcolor", Override: "BGColor"},
				{Name: "span"},
			},
		},
		"data": Desc{
			Attributes: []Attr{
				{Name: "value"},
			},
		},
		"datalist": Desc{
			Override: "DataList",
		},
		"dd": Desc{},
		"del": Desc{
			Attributes: []Attr{
				{Name: "cite"},
				{Name: "datetime", Override: "DateTime"},
			},
		},
		"details": Desc{
			Attributes: []Attr{
				{Name: "open", Type: "bool"},
			},
		},
		"dfn": Desc{},
		"dialog": Desc{
			Attributes: []Attr{
				{Name: "open", Type: "bool"},
			},
		},
		// "div"
		"dl": Desc{},
		"dt": Desc{},
		"em": Desc{},
		"embed": Desc{
			Attributes: []Attr{
				{Name: "height"},
				{Name: "src"},
				{Name: "type"},
				{Name: "width"},
			},
		},
		"fieldset": Desc{
			Attributes: []Attr{
				{Name: "disabled", Type: "bool"},
				{Name: "form"},
				{Name: "name"},
			},
			Override: "FieldSet",
		},
		"figcaption": Desc{
			Override: "FigCaption",
		},
		"figure": Desc{},
		// "footer"
		// "form"
		// "h1"
		"h2": Desc{},
		// "h3"
		// "h4"
		"h5":     Desc{},
		"h6":     Desc{},
		"head":   Desc{},
		"header": Desc{},
		"hgroup": Desc{
			Override: "HGroup",
		},
		// "hr"
		"html": Desc{
			Attributes: []Attr{
				{Name: "xmlns", Override: "XMLNS"},
			},
			Override: "HTML",
		},
		// "i"
		// "iframe"
		// "img"
		// "input"
		"ins": Desc{
			Attributes: []Attr{
				{Name: "cite"},
				{Name: "datetime", Override: "DateTime"},
			},
		},
		"kbd": Desc{},
		// "label"
		"legend": Desc{},
		// "li"
		"link": Desc{
			Attributes: []Attr{
				{Name: "as"},
				{Name: "crossorigin", Override: "CrossOrigin"},
				{Name: "disabled", Type: "bool"},
				{Name: "href"},
				{Name: "hreflang", Override: "HrefLang"},
				{Name: "integrity"},
				{Name: "media"},
				{Name: "methods"},
				{Name: "prefetch"},
				{Name: "referrerpolicy", Override: "ReferrerPolicy"},
				{Name: "rel"},
				{Name: "sizes"},
				{Name: "target"},
				{Name: "title"},
				{Name: "type"},
			},
		},
		"main": Desc{},
		"map": Desc{
			Attributes: []Attr{
				{Name: "name"},
			},
		},
		"mark": Desc{},
		"menu": Desc{
			Attributes: []Attr{
				{Name: "type"},
			},
		},
		"meta": Desc{
			Attributes: []Attr{
				{Name: "charset", Override: "CharSet"},
				{Name: "content"},
				{Name: "http-equiv", Override: "HTTPEquiv"},
				{Name: "name"},
			},
		},
		"meter": Desc{
			Attributes: []Attr{
				{Name: "value", Type: "float64"},
				{Name: "min", Type: "float64"},
				{Name: "max", Type: "float64"},
				{Name: "low", Type: "float64"},
				{Name: "high", Type: "float64"},
				{Name: "optimum", Type: "float64"},
				{Name: "form"},
			},
		},
		// "nav"
		"noscript": Desc{
			Override: "NoScript",
		},
		"object": Desc{
			Attributes: []Attr{
				{Name: "data"},
				{Name: "form"},
				{Name: "height"},
				{Name: "name"},
				{Name: "type"},
				{Name: "typemustmatch", Override: "TypeMustMatch"},
				{Name: "usemap", Override: "UseMap"},
				{Name: "width"},
			},
		},
		"ol": Desc{
			Attributes: []Attr{
				{Name: "compact"},
				{Name: "reversed", Type: "bool"},
				{Name: "start"},
				{Name: "type"},
			},
		},
		"optgroup": Desc{
			Attributes: []Attr{
				{Name: "disabled", Type: "bool"},
				{Name: "label"},
			},
			Override: "OptGroup",
		},
		// "option"
		"output": Desc{
			Attributes: []Attr{
				{Name: "for"},
				{Name: "form"},
				{Name: "name"},
			},
		},
		// "p"
		"param": Desc{
			Attributes: []Attr{
				{Name: "name"},
				{Name: "value"},
			},
		},
		"picture": Desc{},
		// "pre"
		"progress": Desc{
			Attributes: []Attr{
				{Name: "max", Type: "float64"},
				{Name: "value", Type: "float64"},
			},
		},
		"q": Desc{
			Attributes: []Attr{
				{Name: "cite"},
			},
		},
		"rp": Desc{
			Override: "RP",
		},
		"rt": Desc{
			Override: "RT",
		},
		"rtc": Desc{
			Override: "RTC",
		},
		"ruby": Desc{},
		"s": Desc{
			Override: "Strike", // The name is different from <s> because of an identifier name conflict.
		},
		"samp": Desc{},
		"script": Desc{
			Attributes: []Attr{
				{Name: "async"},
				{Name: "crossorigin", Override: "CrossOrigin"},
				{Name: "defer"},
				{Name: "integrity"},
				{Name: "nomodule", Override: "NoModule"},
				{Name: "nonce"},
				{Name: "src"},
				{Name: "text"},
				{Name: "type"},
			},
		},
		"section": Desc{},
		// "select"
		"slot": Desc{
			Attributes: []Attr{
				{Name: "name"},
			},
		},
		"small": Desc{},
		"source": Desc{
			Attributes: []Attr{
				{Name: "sizes"},
				{Name: "src"},
				{Name: "srcset", Override: "SrcSet"},
				{Name: "type"},
				{Name: "media"},
			},
		},
		// "span"
		"strong": Desc{},
		"style": Desc{
			Attributes: []Attr{
				{Name: "type"},
				{Name: "media"},
				{Name: "nonce"},
				{Name: "title"},
			},
		},
		"sub": Desc{},
		// "table"
		"tbody": Desc{
			Attributes: []Attr{
				{Name: "bgcolor", Override: "BGColor"},
			},
		},
		"td": Desc{
			Attributes: []Attr{
				{Name: "bgcolor", Override: "BGColor"},
				{Name: "colspan", Override: "ColSpan"},
				{Name: "headers"},
				{Name: "rowspan", Override: "RowSpan"},
			},
		},
		"template": Desc{},
		"tfoot": Desc{
			Attributes: []Attr{
				{Name: "bgcolor", Override: "BGColor"},
			},
		},
		"th": Desc{
			Attributes: []Attr{
				{Name: "abbr"},
				{Name: "bgcolor", Override: "BGColor"},
				{Name: "colspan", Override: "ColSpan"},
				{Name: "headers"},
				{Name: "rowspan", Override: "RowSpan"},
				{Name: "scope"},
			},
		},
		"thead": Desc{
			Attributes: []Attr{
				{Name: "bgcolor", Override: "BGColor"},
			},
		},
		"time": Desc{
			Attributes: []Attr{
				{Name: "datetime", Override: "DateTime"},
			},
		},
		"title": Desc{},
		"tr":    Desc{},
		"track": Desc{
			Attributes: []Attr{
				{Name: "default", Type: "bool"},
				{Name: "kind"},
				{Name: "label"},
				{Name: "src"},
				{Name: "srclang", Override: "SrcLang"},
			},
		},
		"u": Desc{},
		// "ul"
		"var": Desc{},
		"video": Desc{
			Attributes: []Attr{
				{Name: "autoplay"},
				{Name: "buffered"},
				{Name: "controls"},
				{Name: "crossorigin", Override: "CrossOrigin"},
				{Name: "height"},
				{Name: "loop"},
				{Name: "muted"},
				{Name: "played"},
				{Name: "preload"},
				{Name: "poster"},
				{Name: "src"},
				{Name: "width"},
				{Name: "playsinline", Override: "PlaysInline"},
			},
		},
		"wbr": Desc{},
	}
)

func main() {
	flag.CommandLine.Usage = usage
	flag.Parse()

	primary := template.Must(template.New("primary").Parse(primaryTemplate))
	test := template.Must(template.New("test").Parse(testTemplate))

	for k, v := range elements {
		var upper string
		if v.Override != "" {
			upper = v.Override
		} else {
			upper = strings.ToUpper(string(k[0])) + k[1:]
		}

		var attrs []templAttr
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
			attrs = append(attrs, templAttr{Name: name, JS: js, Type: t})
		}

		e := templElem{
			Elem:  upper + "Elem",
			Name:  k,
			Props: upper + "Props",
			Upper: upper,
			Attrs: attrs,
		}

		executeTemplate(k+"_elem.go", primary, e)
		executeTemplate(k+"_elem_test.go", test, e)
	}
}

func executeTemplate(n string, t *template.Template, e templElem) {
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
