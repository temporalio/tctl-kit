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

package output_test

import (
	"flag"
	"os"

	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

func ExamplePrintCards() {
	ctx, teardown := setupCardTest()
	defer teardown()

	structItems := []*dataForCard{
		{
			Name: "foo1",
			Nested: struct {
				NName  string
				NValue string
			}{
				NName:  "baz1",
				NValue: "qux1",
			},
		},
	}

	var items []interface{}
	for _, item := range structItems {
		items = append(items, item)
	}

	po := output.PrintOptions{
		Fields:     []string{"Name"},
		FieldsLong: []string{"Nested.NName", "Nested.NValue"},
	}

	output.PrintCards(ctx, os.Stdout, items, &po)

	// Output:
	// Name  foo1
}

func setupCardTest() (*cli.Context, func()) {
	app := cli.NewApp()
	flagSet := flag.FlagSet{}
	ctx := cli.NewContext(app, &flagSet, nil)

	return ctx, func() {}
}

type dataForCard struct {
	Name   string
	Nested struct {
		NName  string
		NValue string
	}
}
