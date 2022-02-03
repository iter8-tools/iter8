package base

import (
	"fmt"
	"reflect"
)

// Note: the following code snippets are from sprig library
// https://github.com/Masterminds/sprig

// The following copyright notice is from the sprig library.
// This copyright applies to the code in this file.
// It is included as required by the MIT License under which sprig is released.

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

// Uniq deduplicates a list
// We have switched from uniq to Uniq, since we want to use it in other packages
func Uniq(list interface{}) []interface{} {
	l, err := mustUniq(list)
	if err != nil {
		panic(err)
	}

	return l
}

// mustUniq deduplicates a list and returns an error if the type doesn't permit equality checks
// this function has been modified from the sprig implementation, in order to use the
// the two valued inList function
func mustUniq(list interface{}) ([]interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		dest := []interface{}{}
		var item interface{}
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if ok, _ := inList(dest, item); !ok {
				dest = append(dest, item)
			}
		}

		return dest, nil
	default:
		return nil, fmt.Errorf("cannot find uniq on type %s", tp)
	}
}

// inList checks if needle is present in haystack
// this function has been modified from the sprig implementation, in order to return index also
func inList(haystack []interface{}, needle interface{}) (bool, int) {
	for i, h := range haystack {
		if reflect.DeepEqual(needle, h) {
			return true, i
		}
	}
	return false, -1
}
