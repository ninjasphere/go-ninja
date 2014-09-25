package ninja

import "regexp"

var params, _ = regexp.Compile(":[^/$]+")

func GetSubscribeTopic(topic string) string {
	return params.ReplaceAllString(topic, "+")
}

// Adapted From: https://github.com/bmizerany/pat
/*
Copyright (C) 2012 by Keith Rarick, Blake Mizerany

Permission is hereby granted, free of charge, to any person obtaining a copy of this
software and associated documentation files (the "Software"), to deal in the Software
without restriction, including without limitation the rights to use, copy, modify, merge,
publish, distribute, sublicense, and/or sell copies of the Software, and to permit
persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT
OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

*/

func MatchTopicPattern(pattern, path string) (*map[string]string, bool) {
	p := make(map[string]string)
	var i, j int
	for i < len(path) {
		switch {
		case j >= len(pattern):
			if pattern != "/" && len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
				return &p, true
			}
			return nil, false
		case pattern[j] == ':':
			var name, val string
			var nextc byte
			name, nextc, j = match(pattern, isAlnum, j+1)
			val, _, i = match(path, matchPart(nextc), i)
			p[name] = val
		case path[i] == pattern[j]:
			i++
			j++
		default:
			return nil, false
		}
	}
	if j != len(pattern) {
		return nil, false
	}
	return &p, true
}

func matchPart(b byte) func(byte) bool {
	return func(c byte) bool {
		return c != b && c != '/'
	}
}

func match(s string, f func(byte) bool, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) {
		j++
	}
	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}
