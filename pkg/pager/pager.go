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

package pager

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

const (
	DefaultListPageSize = 20
)

// NewPager returns a writer such as stdout, "less", "more" or a pager provided by the user.
// A user can provide the pager name with a pager flag or env variable.
// If no pager is provided, it will fall back to stdout.
func NewPager(c *cli.Context, pager string) (io.Writer, func()) {
	noPager := c.Bool(FlagNoPager)
	if noPager || pager == "" || pager == string(Stdout) {
		return os.Stdout, func() {}
	}

	exe, err := lookupPager(pager)
	if err != nil {
		return os.Stdout, func() {}
	}

	cmd := exec.Command(exe)

	if pager == string(Less) {
		env := os.Environ()
		env = append(env, "LESS=FRX")
		cmd.Env = env
	}

	signal.Ignore(syscall.SIGPIPE)

	reader, writer := io.Pipe()
	cmd.Stdin = reader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}()

	return writer, func() {
		writer.Close()
		<-done
	}
}

func lookupPager(pagerName string) (string, error) {
	if pagerName == "" {
		return "", errors.New("no pager provided")
	}

	if path, err := exec.LookPath(pagerName); err == nil {
		return path, nil
	}

	return pagerName, errors.New("unable to find pager " + pagerName)
}
