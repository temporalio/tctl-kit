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

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

type Config struct {
	viper *viper.Viper
}

func NewConfig(appName, configName string) (*Config, error) {
	dpath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dpath = filepath.Join(dpath, ".config", appName)

	mkfile(dpath, configName+".yaml")

	v := viper.New()
	v.AddConfigPath(dpath)
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.SetDefault("aliases", map[string]string{})
	v.SetDefault("active", "local")
	v.SetDefault("environments", map[string]interface{}{"local": interface{}("")})

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			return nil, err
		}
	}

	return &Config{viper: v}, nil
}

func (c *Config) GetByEnvironment(ctx *cli.Context, key string) (string, error) {
	activeEnv := c.viper.GetString("active")
	fullKey := activeEnv + "." + key

	if c.viper.IsSet(fullKey) {
		return c.viper.GetString(fullKey), nil
	}

	return "", nil
}

func (c *Config) SetByEnvironment(ctx *cli.Context, key string, value string) error {
	activeEnv := c.viper.GetString("active")
	if activeEnv == "" {
		return fmt.Errorf("active environment not set")
	}

	fullKey := activeEnv + "." + key
	c.viper.Set(fullKey, value)

	if err := c.viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

func (c *Config) GetGlobal(ctx *cli.Context, key string) (string, error) {
	if c.viper.IsSet(key) {
		return c.viper.GetString(key), nil
	}

	return "", nil
}

func (c *Config) SetGlobal(ctx *cli.Context, key string, value string) error {
	c.viper.Set(key, value)

	if err := c.viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

func mkfile(dirPath, filename string) (string, error) {
	err := mkdir(dirPath)
	if err != nil {
		return "", err
	}
	fpath := filepath.Join(dirPath, filename)

	if _, err := os.Stat(fpath); err != nil {
		fmt.Printf("creating config file: %v\n", fpath)
		file, err := os.Create(fpath)
		if err != nil {
			defer file.Close()
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
