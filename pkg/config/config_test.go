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

package config_test

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/temporalio/tctl-kit/pkg/config"
)

const (
	appName = "test-tctl-kit"
)

func TestNewConfigPermissionDenied(t *testing.T) {
	expectedError := errors.New("open /tmp.yaml: permission denied")

	cfg, err := config.NewConfig(appName, "../../../../../../../../../../../../tmp")
	assert.NoError(t, err)

	err = cfg.SetEnvProperty("test", "test", "test")

	if assert.Error(t, err) {
		assert.Equal(t, expectedError.Error(), err.Error())
	}
}

func TestCreatesConfigFileLazily(t *testing.T) {
	cfg, teardown := setupConfig(t, "")
	defer teardown()

	path := cfg.Path()

	_, err := os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)

	cfg.SetEnvProperty("test", "test", "test")

	_, err = os.Stat(path)
	assert.NoError(t, err)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		assert.Error(t, err)
	}
}

type envprop struct {
	env, key, value string
}

func TestAlias(t *testing.T) {
	testcases := map[string]struct {
		input  string
		expect map[string]string
	}{
		"reads alias by name": {
			input: `aliases:
    wj: workflow list --output json
    wt: workflow list --output table`,
			expect: map[string]string{
				"wj":           "workflow list --output json",
				"wt":           "workflow list --output table",
				"doesnt-exist": "",
				"":             "",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, tc.input)
			defer teardown()

			for aliasName, vExpected := range tc.expect {
				vActual := cfg.Alias(aliasName)
				assert.Equal(t, vActual, vExpected)
			}
		})
	}
}

func TestSetAlias(t *testing.T) {
	testcases := map[string]struct {
		input    map[string]string
		expected string
		err      bool
	}{
		"throws on empty key": {
			input: map[string]string{
				"": "value",
			},
			err: true,
		},
		"throws on invalid key": {
			input: map[string]string{
				"key!": "value",
				"-key": "value",
				"k.ey": "value",
			},
			err: true,
		},
		"sets on proper key and value ": {
			input: map[string]string{
				"wt": "workflow list --output table",
				"wj": "workflow list --output json",
			},
			expected: "aliases:\n    wj: workflow list --output json\n    wt: workflow list --output table",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, "")

			if !tc.err {
				defer teardown()
			}

			for key, value := range tc.input {
				err := cfg.SetAlias(key, value)
				if tc.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if !tc.err {
				assert.Contains(t, readConfig(t, cfg), tc.expected)
			}
		})
	}
}

func TestCurrentEnv(t *testing.T) {
	testcases := map[string]struct {
		input  string
		expect map[string]string
	}{
		"reads alias by name": {
			input: `aliases:
    wj: workflow list --output json
    wt: workflow list --output table`,
			expect: map[string]string{
				"wj":           "workflow list --output json",
				"wt":           "workflow list --output table",
				"doesnt-exist": "",
				"":             "",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, tc.input)
			defer teardown()

			for aliasName, vExpected := range tc.expect {
				vActual := cfg.Alias(aliasName)
				assert.Equal(t, vActual, vExpected)
			}
		})
	}
}

func TestSetCurrentEnv(t *testing.T) {
	testcases := map[string]struct {
		input    string
		expected string
		err      bool
	}{
		"throws on empty value": {
			input: "",
			err:   true,
		},
		"throws on invalid value": {
			input: "wrong-env-name!",
			err:   true,
		},
		"sets on proper value ": {
			input:    "dev",
			expected: "current-env: dev",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, "")
			defer teardown()

			err := cfg.SetCurrentEnv(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, readConfig(t, cfg), tc.expected)
			}
		})
	}
}

func TestEnv(t *testing.T) {
	testcases := map[string]struct {
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

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, tc.input)
			defer teardown()

			for envName, envProps := range tc.expect {
				envActual := cfg.Env(envName)
				for key, vExpected := range envProps {
					vActual := envActual[key]
					assert.Equal(t, vActual, vExpected)
				}
			}
		})
	}
}

func TestRemoveEnv(t *testing.T) {
	testcases := map[string]struct {
		inputCfg    string
		inputRemove string
		expected    string
		err         bool
	}{
		"throws on empty name": {
			inputRemove: "",
			err:         true,
		},
		"throws on invalid name": {
			inputRemove: "wrong-env-name!",
			err:         true,
		},
		"throws on removing current env": {
			inputCfg:    "current-env: dev",
			inputRemove: "dev",
			err:         true,
		},
		"removes env on proper name": {
			inputCfg: `
current-env: remote
env:
  local:
    key1: value-local-1
    key2: value-local-2
  remote:
    key1: value-remote-1`,
			inputRemove: "local",
			expected: `
current-env: remote
env:
    remote:
        key1: value-remote-1`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, tc.inputCfg)
			defer teardown()

			err := cfg.RemoveEnv(tc.inputRemove)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, readConfig(t, cfg), tc.expected)
			}
		})
	}
}

func TestEnvProperty(t *testing.T) {
	testcases := map[string]struct {
		input  string
		expect map[string]map[string]string
	}{
		"reads env property by key": {
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

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, tc.input)
			defer teardown()

			for envName, envProps := range tc.expect {
				for key, vExpected := range envProps {
					vActual, err := cfg.EnvProperty(envName, key)
					assert.NoError(t, err)
					assert.Equal(t, vActual, vExpected)
				}
			}
		})
	}

}

func TestSetEnvProperty(t *testing.T) {
	testcases := map[string]struct {
		input    []envprop
		expected string
		err      bool
	}{
		"throws on empty key": {
			input: []envprop{{"", "key", "value"}, {"local", "", "value"}},
			err:   true,
		},
		"throws on invalid key": {
			input: []envprop{
				{"1local", "key", "value"},
				{"local", "1key", "value"},
				{"-local", "key", "value"},
				{"local", "-key", "value"},
				{"local!", "key", "value"},
				{"local", "key!", "value"},
				{"lo.cal", "key", "value"},
				{"local", "ke.y", "value"},
			},
			err: true,
		},
		"accepts empty value": {
			input:    []envprop{{"local", "key", ""}},
			expected: "local:\n        key: \"\"",
			err:      false,
		},
		"merges env properties": {
			input: []envprop{
				{"local", "key1", "value-local-1"},
				{"local", "key2", "value-local-2"},
				{"remote", "key3", "value-remote-3"},
			},
			expected: `local:
        key1: value-local-1
        key2: value-local-2
    remote:
        key3: value-remote-3`,
			err: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, "")
			if !tc.err {
				defer teardown()
			}

			for _, envprop := range tc.input {
				err := cfg.SetEnvProperty(envprop.env, envprop.key, envprop.value)
				if tc.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if !tc.err {
				assert.Contains(t, readConfig(t, cfg), tc.expected)
			}
		})
	}
}

func setupConfig(t *testing.T, content string) (*config.Config, func()) {
	file := "config-" + uuid.New()[:4] + ".yaml"

	cfg, err := config.NewConfig(appName, file)
	assert.NoError(t, err)

	if content != "" {
		writeConfig(t, cfg.Path(), content)
		cfg, err = config.NewConfig(appName, file)
		assert.NoError(t, err)
	}

	teardown := func() {
		path := cfg.Path()

		err := os.Remove(path)
		assert.NoError(t, err)
	}

	return cfg, teardown
}

func readConfig(t *testing.T, cfg *config.Config) string {
	path := cfg.Path()
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func writeConfig(t *testing.T, path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
