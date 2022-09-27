// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v2"
)

type cardColumns struct {
	Name  string
	Value interface{}
}

func PrintCards(c *cli.Context, w io.Writer, items []interface{}, opts *PrintOptions) error {
	fields := opts.Fields
	if fields == nil {
		fields = extractFieldNames(items[0], []string{}, "", fieldsDepth)
	}

	valuesList, err := extractFieldValues(items, fields)
	if err != nil {
		return fmt.Errorf("unable to print card view: %w", err)
	}

	for _, obj := range valuesList {
		var rows []*cardColumns

		for j, fieldValue := range obj {
			rows = append(rows, &cardColumns{
				Name:  fields[j],
				Value: fieldValue,
			})
		}

		var rowsI []interface{}
		for _, row := range rows {
			rowsI = append(rowsI, row)
		}

		opts.NoHeader = true
		opts.Fields = []string{"Name", "Value"}
		err = PrintTable(c, w, rowsI, opts)
		if err != nil {
			return err
		}

		fmt.Fprintln(w, strings.Repeat(opts.Separator, 10))
	}

	return nil
}
