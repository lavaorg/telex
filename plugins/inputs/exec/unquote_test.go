package exec

/*
Copyright (C) 2014 Kevin Ballard

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
	"reflect"
	"testing"
)

func TestSimpleSplit(t *testing.T) {
	for _, elem := range simpleSplitTest {
		output, err := Split(elem.input)
		if err != nil {
			t.Errorf("Input %q, got error %#v", elem.input, err)
		} else if !reflect.DeepEqual(output, elem.output) {
			t.Errorf("Input %q, got %q, expected %q", elem.input, output, elem.output)
		}
	}
}

func TestErrorSplit(t *testing.T) {
	for _, elem := range errorSplitTest {
		_, err := Split(elem.input)
		if err != elem.error {
			t.Errorf("Input %q, got error %#v, expected error %#v", elem.input, err, elem.error)
		}
	}
}

var simpleSplitTest = []struct {
	input  string
	output []string
}{
	{"hello", []string{"hello"}},
	{"hello goodbye", []string{"hello", "goodbye"}},
	{"hello   goodbye", []string{"hello", "goodbye"}},
	{"glob* test?", []string{"glob*", "test?"}},
	{"don\\'t you know the dewey decimal system\\?", []string{"don't", "you", "know", "the", "dewey", "decimal", "system?"}},
	{"'don'\\''t you know the dewey decimal system?'", []string{"don't you know the dewey decimal system?"}},
	{"one '' two", []string{"one", "", "two"}},
	{"text with\\\na backslash-escaped newline", []string{"text", "witha", "backslash-escaped", "newline"}},
	{"text \"with\na\" quoted newline", []string{"text", "with\na", "quoted", "newline"}},
	{"\"quoted\\d\\\\\\\" text with\\\na backslash-escaped newline\"", []string{"quoted\\d\\\" text witha backslash-escaped newline"}},
	{"text with an escaped \\\n newline in the middle", []string{"text", "with", "an", "escaped", "newline", "in", "the", "middle"}},
	{"foo\"bar\"baz", []string{"foobarbaz"}},
}

var errorSplitTest = []struct {
	input string
	error error
}{
	{"don't worry", UnterminatedSingleQuoteError},
	{"'test'\\''ing", UnterminatedSingleQuoteError},
	{"\"foo'bar", UnterminatedDoubleQuoteError},
	{"foo\\", UnterminatedEscapeError},
	{"   \\", UnterminatedEscapeError},
}
