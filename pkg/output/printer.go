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
	// Fields is a list of fields to print
	Fields []string
	// FieldsLong is a list of additional fields to print with "--fields long" flag
	FieldsLong []string
	// ForceFields ignores user provided fields and uses print options instead. Useful when printing secondary data
	ForceFields bool
	// OutputFormat is the output format to use: table, json..
	OutputFormat OutputOption
	// Pager is the pager to use for interactive mode. Default - stdout
	Pager pager.PagerOption
	// NoHeader removes the header in the table output
	NoHeader bool
	// Separator to use in table output
	Separator string
}

// PrintItems prints items based on user flags or print options.
func PrintItems(c *cli.Context, items []interface{}, opts *PrintOptions) {
	fields := c.String(FlagFields)

	pagerName := c.String(pager.FlagPager)
	if pagerName == "" {
		pagerName = string(opts.Pager)
	}

	writer, close := pager.NewPager(c, pagerName)
	defer close()

	if !opts.ForceFields && c.IsSet(FlagFields) {
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

	output := getOutputFormat(c, opts)
	switch output {
	case Table:
		PrintTable(c, writer, items, opts)
	case JSON:
		PrintJSON(c, writer, items)
	case Card:
		PrintCards(c, writer, items, opts)
	default:
	}
}

// PrintIterator prints items from an iterator based on user flags or print options.
func PrintIterator(c *cli.Context, iter iterator.Iterator, opts *PrintOptions) error {
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

func getOutputFormat(c *cli.Context, opts *PrintOptions) OutputOption {
	outputFlag := c.String(FlagOutput)
	output := OutputOption(outputFlag)

	if opts != nil {
		if c.IsSet(FlagOutput) {
			return output
		} else if opts.OutputFormat != "" {
			return opts.OutputFormat
		}
	}

	return Table
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
