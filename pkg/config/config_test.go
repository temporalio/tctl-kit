// The MIT License
//
// Copyright (c) 2021 Temporal Technologies Inc.  All rights reserved.
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

package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	appName = "test-tctl-kit"
	cfgFile = "test-tctl-kit-config"
)

func TestNewConfigCreatesFile(t *testing.T) {
	path := getConfigPath(t)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		assert.Error(t, err) // config file shouldn't exist yet
	}

	createConfig(t)

	_, err := os.Stat(path)
	assert.NoError(t, err)

	removeConfig(t)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		assert.Error(t, err) // config file shouldn't exist
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := createConfig(t)
	defer removeConfig(t)

	env, err := cfg.Get(KeyCurrentEnvironment)
	assert.NoError(t, err)
	assert.Equal(t, env, "local")
}

func TestConfigSetKey(t *testing.T) {
	tests := map[string]struct {
		keyvalues map[string]string
		err       bool
	}{
		"throw on empty key": {
			keyvalues: map[string]string{
				"": "value",
			},
			err: true,
		},
		"throw on invalid key": {
			keyvalues: map[string]string{
				"1key": "value",
				"-key": "value",
				"key!": "value",
			},
			err: true,
		},
		"accepts empty value": {
			keyvalues: map[string]string{
				"key": "",
			},
			err: false,
		},
		"valid key and value": {
			keyvalues: map[string]string{
				"valid-key":                  "value 1",
				"valid-key2":                 "value 2",
				"valid-key3.xxx-yyy_zzz.ooo": "value 3",
			},
			err: false,
		},
		"merge keys": {
			keyvalues: map[string]string{
				"env.local.key1":  "value-local-1",
				"env.local.key2":  "value-local-2",
				"env.remote.key1": "value-remote-1",
			},
			err: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := createConfig(t)
			defer removeConfig(t)

			for key, value := range tc.keyvalues {
				err := cfg.Set(key, value)
				if tc.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if !tc.err {
				for key, value := range tc.keyvalues {
					v, err := cfg.Get(key)
					assert.NoError(t, err)
					assert.Equal(t, v, value)
				}
			}
		})
	}
}

func createConfig(t *testing.T) *Config {
	cfg, err := NewConfig(appName, cfgFile)
	assert.NoError(t, err)

	return cfg
}

func removeConfig(t *testing.T) {
	path := getConfigPath(t)

	err := os.Remove(path)
	assert.NoError(t, err)
}

func getConfigPath(t *testing.T) string {
	dpath, err := os.UserHomeDir()
	assert.NoError(t, err)

	path := filepath.Join(dpath, ".config", appName, cfgFile+".yaml")

	return path
}
