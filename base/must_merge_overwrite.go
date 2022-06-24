/*
Copyright (C) 2013-2020 Masterminds

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package base

import (
	"github.com/imdario/mergo"
	"github.com/iter8-tools/iter8/base/log"
)

// mustMergeOverwrite	merge maps giving precedence to the right side
func mustMergeOverwrite(dst map[string]interface{}, srcs ...map[string]interface{}) (interface{}, error) {
	for _, src := range srcs {
		if err := mergo.MergeWithOverwrite(&dst, src); err != nil {
			// the following log line is the diff between the original sprig func and ours
			log.Logger.Error(err)
			return nil, err
		}
	}
	return dst, nil
}
