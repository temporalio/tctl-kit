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
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/temporalio/tctl-kit/pkg/format"
	"github.com/temporalio/tctl-kit/pkg/iterator"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
)

const (
	BatchPrintSize = 100
)

type PrintOptions struct {
	Fields      []string
	FieldsLong  []string
	IgnoreFlags bool
	Output      OutputOption
	Pager       io.Writer
	NoPager     bool
	NoHeader    bool
	Separator   string
}

// PrintItems prints items based on user flags or print options.
// User flags are prioritized unless IgnoreFlags is set
func PrintItems(c *cli.Context, items []interface{}, opts *PrintOptions) {
	fields := c.String(FlagFields)

	output := getOutputFormat(c, opts)
	pager, close := newPager(c, opts)
	opts.Pager = pager
	defer close()

	if !opts.IgnoreFlags && c.IsSet(FlagFields) {
		if fields == FieldsLong {
			opts.Fields = append(opts.Fields, opts.FieldsLong...)
			opts.FieldsLong = []string{}
		} else {
			f := strings.Split(fields, ",")
			for i := range f {
				f[i] = strings.TrimSpace(f[i])
			}
			opts.Fields = f
			opts.FieldsLong = []string{}
		}
	}

	switch output {
	case Table:
		PrintTable(c, items, opts)
	case JSON:
		PrintJSON(c, items, opts)
	case Card:
		PrintCards(c, items, opts)
	default:
	}
}

// Pager creates an interactive CLI mode to control the printing of items
func Pager(c *cli.Context, iter iterator.Iterator, opts *PrintOptions) error {
	limit := c.Int(FlagLimit)

	if opts == nil {
		opts = &PrintOptions{}
	}

	itemsPrinted := 0
	var batch []interface{}
	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			return err
		}

		if c.IsSet(FlagLimit) && itemsPrinted >= limit {
			break
		}

		batch = append(batch, item)
		itemsPrinted++

		follow := c.Bool(FlagFollow)
		isLastBatch := limit-itemsPrinted < BatchPrintSize
		isBatchFilled := (len(batch) == BatchPrintSize) || (isLastBatch && len(batch) == limit%BatchPrintSize)

		if follow || isBatchFilled || !iter.HasNext() {
			// for consistent formatting, print items in batches (ex. in Table output)
			// else if --follow is on, print items as they are received
			PrintItems(c, batch, opts)
			batch = batch[:0]
			opts.NoHeader = true
		}
	}

	return nil
}

// newPager creates a new pager based on user flags or print options
// User flags are prioritized unless IgnoreFlags is set
func newPager(c *cli.Context, opts *PrintOptions) (io.Writer, func()) {
	if opts.NoPager || c.Bool(pager.FlagNoPager) {
		return os.Stdout, func() {}
	}

	output := getOutputFormat(c, opts)
	suggestedPager := suggestPagerByOutputFormat(c, output)

	return pager.NewPager(c, suggestedPager)
}

func getOutputFormat(c *cli.Context, opts *PrintOptions) OutputOption {
	outputFlag := c.String(FlagOutput)
	output := OutputOption(outputFlag)

	if opts != nil {
		if !opts.IgnoreFlags && c.IsSet(FlagOutput) {
			return output
		} else if opts.Output != "" {
			return opts.Output
		}
	}

	return Table
}

// suggestPagerByOutputFormat returns the suggested pager by output format
// For Table and Card views suggests 'less' as the pager
// For JSON, as it tends to be larger, suggests 'more' as the pager
func suggestPagerByOutputFormat(c *cli.Context, oo OutputOption) string {
	switch oo {
	case "":
		return string(pager.Stdout)
	case Table:
		return string(pager.Less)
	case Card:
		return string(pager.Less)
	case JSON:
		return string(pager.More)
	default:
		return string(pager.Stdout)
	}
}

func formatField(c *cli.Context, i interface{}) string {
	val := reflect.ValueOf(i)
	val = reflect.Indirect(val)

	var typ reflect.Type
	if val.IsValid() && !val.IsZero() {
		typ = val.Type()
	}
	kin := val.Kind()

	if typ == reflect.TypeOf(time.Time{}) {
		return format.FormatTime(c, val.Interface().(time.Time))
	} else if kin == reflect.Struct && val.CanInterface() {
		str, _ := ParseToJSON(c, i, false)

		return str
	} else {
		return fmt.Sprintf("%v", i)
	}
}
