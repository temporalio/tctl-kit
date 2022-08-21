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
	"log"
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
		assert.Error(t, err)
	}

	initConfig(t)

	_, err := os.Stat(path)
	assert.NoError(t, err)

	removeConfig(t)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		assert.Error(t, err)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := initConfig(t)
	defer removeConfig(t)

	env, err := cfg.Get(KeyCurrentEnvironment)
	assert.NoError(t, err)
	assert.Equal(t, env, "local")
}

func TestConfigSetKey(t *testing.T) {
	tests := map[string]struct {
		input    map[string]string
		expected string
		err      bool
	}{
		"throw on empty key": {
			input: map[string]string{
				"": "value",
			},
			err: true,
		},
		"throw on invalid key": {
			input: map[string]string{
				"1key": "value",
				"-key": "value",
				"key!": "value",
			},
			err: true,
		},
		"accepts empty value": {
			input: map[string]string{
				"key": "",
			},
			expected: "env:\n  local: {}\nkey: \"\"\n",
			err:      false,
		},
		"valid key and value": {
			input: map[string]string{
				"valid-key":                  "value 1",
				"valid-key2":                 "value 2",
				"valid-key3.xxx-yyy_zzz.ooo": "value 3",
			},
			expected: `valid-key: value 1
valid-key2: value 2
valid-key3:
  xxx-yyy_zzz:
    ooo: value 3
`,
			err: false,
		},
		"merge keys": {
			input: map[string]string{
				"env.local.key1":  "value-local-1",
				"env.local.key2":  "value-local-2",
				"env.remote.key1": "value-remote-1",
			},
			expected: `  local:
    key1: value-local-1
    key2: value-local-2
  remote:
    key1: value-remote-1`,
			err: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := initConfig(t)
			defer removeConfig(t)

			for key, value := range tc.input {
				err := cfg.Set(key, value)
				if tc.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if !tc.err {
				assert.Contains(t, readConfig(t), tc.expected)
			}
		})
	}
}

func TestConfigGetEnv(t *testing.T) {
	tests := map[string]struct {
		input  string
		expect map[string]map[string]string
	}{
		"reads env by name": {
			input: `env:
  local:
    key1: value-local-1
    key2: value-local-2
  remote:
    key1: value-remote-1`,
			expect: map[string]map[string]string{
				"local": {
					"key1": "value-local-1",
					"key2": "value-local-2",
				},
				"remote": {
					"key1": "value-remote-1",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			writeConfig(t, tc.input)
			cfg := initConfig(t)
			defer removeConfig(t)

			for envName, envProps := range tc.expect {
				envActual, err := cfg.GetEnv(envName)
				assert.NoError(t, err)
				for key, vExpected := range envProps {
					vActual := envActual[key]
					assert.Equal(t, vActual, vExpected)
				}
			}
		})
	}
}

func initConfig(t *testing.T) *Config {
	cfg, err := NewConfig(appName, cfgFile)
	assert.NoError(t, err)

	return cfg
}

func readConfig(t *testing.T) string {
	path := getConfigPath(t)
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func writeConfig(t *testing.T, content string) {
	path := getConfigPath(t)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getConfigPath(t *testing.T) string {
	dpath, err := os.UserHomeDir()
	assert.NoError(t, err)

	path := filepath.Join(dpath, ".config", appName, cfgFile+".yaml")

	return path
}

func removeConfig(t *testing.T) {
	path := getConfigPath(t)

	err := os.Remove(path)
	assert.NoError(t, err)
}
