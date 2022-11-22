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
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

const DefaultEnv = "default"

type Config struct {
	Aliases    map[string]string            `yaml:"aliases"`
	CurrentEnv string                       `yaml:"current-env"`
	Envs       map[string]map[string]string `yaml:"env"`
	Version    string                       `yaml:"version"`

	path string
}

func NewConfig(appName, configName string) (*Config, error) {
	dpath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dpath = filepath.Join(dpath, ".config", appName)

	mkfile(dpath, configName+".yaml")

	cfgPath := filepath.Join(dpath, configName+".yaml")
	cfg, err := readConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	if cfg.CurrentEnv == "" {
		cfg.CurrentEnv = DefaultEnv
	}

	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}

	if cfg.Envs == nil {
		cfg.Envs = map[string]map[string]string{DefaultEnv: {}}
	}

	cfg.path = cfgPath

	return cfg, nil
}

func (c *Config) Alias(name string) string {
	return c.Aliases[name]
}

func (c *Config) SetAlias(name, value string) error {
	if err := validateKey(name); err != nil {
		return fmt.Errorf("invalid alias name: %w", err)
	}

	c.Aliases[name] = value

	return c.writeFile()
}

func (c *Config) SetCurrentEnv(name string) error {
	if err := validateKey(name); err != nil {
		return fmt.Errorf("invalid env name: %w", err)
	}

	c.CurrentEnv = name

	return c.writeFile()
}

func (c *Config) Env(name string) map[string]string {
	return c.Envs[name]
}

func (c *Config) RemoveEnv(name string) error {
	if err := validateKey(name); err != nil {
		return fmt.Errorf("invalid env name: %w", err)
	}

	if c.CurrentEnv == name {
		return fmt.Errorf("unable to remove current env")
	}

	delete(c.Envs, name)

	return c.writeFile()
}

func (c *Config) EnvProperty(env, key string) (string, error) {
	if env, ok := c.Envs[env]; ok {
		return env[key], nil
	}

	return "", fmt.Errorf("env not found: %v", env)
}

func (c *Config) SetEnvProperty(env, key, value string) error {
	if err := validateKey(env); err != nil {
		return fmt.Errorf("invalid env name: %w", err)
	}

	if err := validateKey(key); err != nil {
		return fmt.Errorf("invalid property key: %w", err)
	}

	if _, ok := c.Envs[env]; !ok {
		c.Envs[env] = map[string]string{}
	}

	c.Envs[env][key] = value

	return c.writeFile()
}

func (c *Config) RemoveEnvProperty(envName, key string) error {
	if env, ok := c.Envs[envName]; ok {
		delete(env, key)

		return c.writeFile()
	}

	return nil
}

func (c *Config) SetVersion(value string) error {
	c.Version = value

	return c.writeFile()
}

func mkfile(dirPath, filename string) (string, error) {
	if err := mkdir(dirPath); err != nil {
		return "", err
	}

	fpath := filepath.Join(dirPath, filename)

	if _, err := os.Stat(fpath); err != nil {
		fmt.Printf("creating config file: %v\n", fpath)
		file, err := os.Create(fpath)
		if err == nil {
			defer file.Close()
		} else {
			return fpath, err
		}
	}

	return fpath, nil
}

func mkdir(path string) error {
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("creating config dir: %v\n", path)
		err = os.MkdirAll(path, 0755)
		return err
	}

	return nil
}

func readConfig(path string) (*Config, error) {
	cfgYaml, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(cfgYaml, &config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file: %w", err)
	}

	return &config, nil
}

func (c *Config) writeFile() error {
	cfgYaml, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	err = os.WriteFile(c.path, cfgYaml, 0644)
	if err != nil {
		return fmt.Errorf("unable to write config file: %w", err)
	}

	return nil
}

// validateKey validates the key against the following rules:
// 1. key must start with a letter
// 2. key must contain only word characters and dashes
// 3. key must end with a letter or number
func validateKey(key string) error {
	pattern := `^[a-z][\w\-]*[a-z0-9]$`

	matched, err := regexp.MatchString(pattern, key)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("invalid key: %v. Key must follow pattern: %v", key, pattern)
	}

	return nil
}
