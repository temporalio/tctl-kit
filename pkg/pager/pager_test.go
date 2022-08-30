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

package pager_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
)

func setupPagerTest() (*cli.Context, func()) {
	app := cli.NewApp()
	flagSet := flag.FlagSet{}
	ctx := cli.NewContext(app, &flagSet, nil)

	return ctx, func() {}
}

func TestPrintTable_Stdout(t *testing.T) {
	ctx, teardown := setupPagerTest()
	defer teardown()

	w, _ := pager.NewPager(ctx, "")
	assert.Equal(t, w, os.Stdout)

	w, _ = pager.NewPager(ctx, "stdout")
	assert.Equal(t, w, os.Stdout)
}

func TestPrintTable_StdoutFallback(t *testing.T) {
	ctx, teardown := setupPagerTest()
	defer teardown()

	w, close := pager.NewPager(ctx, "executable-that-doesnt-exist")
	defer close()

	assert.Equal(t, w, os.Stdout)
}

func TestPrintTable_Less(t *testing.T) {
	ctx, teardown := setupPagerTest()
	defer teardown()

	w, close := pager.NewPager(ctx, "less")
	defer close()

	assert.NotEqual(t, w, os.Stdout)
}
