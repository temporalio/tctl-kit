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
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/urfave/cli/v2"
)

func PrintJSON(c *cli.Context, w io.Writer, o interface{}) error {
	json, err := ParseToJSON(o, true)
	if err != nil {
		return fmt.Errorf("unable to print json: %s", err)
	}

	_, err = fmt.Fprintln(w, json)
	return err
}

func ParseToJSON(o interface{}, indent bool) (string, error) {
	var b []byte
	var err error

	if pb, ok := o.(proto.Message); ok {
		encoder := jsonpb.Marshaler{}
		if indent {
			encoder.Indent = "  "
		}

		var buf bytes.Buffer
		err = encoder.Marshal(&buf, pb)
		b = buf.Bytes()
	} else {
		if indent {
			b, err = json.MarshalIndent(o, "", "  ")
		} else {
			b, err = json.Marshal(o)
		}
	}

	if err != nil {
		return "", err
	}

	return string(b), nil
}
