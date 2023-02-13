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
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/temporalio/tctl-kit/pkg/config"
)

const (
	appName = "test-tctl-kit"
)

func TestNewConfigPermissionDenied(t *testing.T) {
	dir, err := os.UserHomeDir()
	assert.NoError(t, err)

	appName := uuid.New()
	readOnly := os.FileMode(0400)
	dir = filepath.Join(dir, ".config", appName)
	os.MkdirAll(dir, readOnly)

	// umask may have changed config folder permissions, ensure they are correct
	err = os.Chmod(dir, os.FileMode(readOnly))
	assert.NoError(t, err)

	_, err = config.NewConfig(appName, "test")

	switch runtime.GOOS {
	case "windows":
		t.Skip("no permission denied error on Windows")
	case "darwin":
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("expected error %T, got %T", fs.ErrPermission, err)
		}
	case "linux":
		if !errors.Is(err, os.ErrPermission) {
			t.Errorf("expected error %v, got %T", fs.ErrPermission, err)
		}
	default:
		t.Errorf("unexpected OS %s", runtime.GOOS)
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

func TestFilePermissionIsOwnerReadWrite(t *testing.T) {
	cfg, teardown := setupConfig(t, "")
	defer teardown()

	// ensure the config file is created
	err := cfg.SetEnvProperty("test", "test", "test")
	assert.NoError(t, err)

	path := cfg.Path()

	fileInfo, err := os.Stat(path)
	assert.NoError(t, err)

	assert.Equal(t, os.FileMode(0600).String(), fileInfo.Mode().String())
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
		"removes env on proper name": {
			inputCfg: `
env:
  local:
    key1: value-local-1
    key2: value-local-2
  remote:
    key1: value-remote-1`,
			inputRemove: "local",
			expected: `env:
    remote:
        key1: value-remote-1`,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg, teardown := setupConfig(t, tc.inputCfg)
			if tc.err && tc.inputCfg != "" {
				defer teardown()
			}

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
	file := "config-" + uuid.New()[:4]

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
