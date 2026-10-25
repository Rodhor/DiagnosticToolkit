package app

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	Timeout     string            `yaml:"timeout"`
	Concurrency *int              `yaml:"concurrency"`
	Thresholds  map[string]string `yaml:"thresholds"`
	Output      *outputFile       `yaml:"output"`
	Monitor     *monitoringFile   `yaml:"monitor"`
	Ping        *pingFile         `yaml:"ping"`
	Targets     *targetsFile      `yaml:"targets"`
	Queries     []queryFile       `yaml:"dns_queries"`
}

type monitoringFile struct {
	DefaultInterval string `yaml:"default_interval"`
}

type pingFile struct {
	Count       *int     `yaml:"count"`
	Interval    string   `yaml:"interval"`
	Timeout     string   `yaml:"timeout"`
	WarnLossPct *float64 `yaml:"warn_loss_pct"`
	FailLossPct *float64 `yaml:"fail_loss_pct"`
	Privileged  *bool    `yaml:"privileged"`
}

type targetsFile struct {
	TCP       []string       `yaml:"tcp"`
	Ping      []string       `yaml:"ping"`
	Resolvers []resolverFile `yaml:"dns_resolvers"`
}

type resolverFile struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type queryFile struct {
	Name     string   `yaml:"name"`
	Expect   string   `yaml:"expect"`
	Expected []string `yaml:"expected"`
}

type outputFile struct {
	Results string `yaml:"results"`
	Summary string `yaml:"summary"`
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	cfg, err := parseConfig(f, Default())
	if err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}

	return cfg, nil
}

func parseConfig(r io.Reader, base Config) (Config, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	var fc fileConfig
	if err := dec.Decode(&fc); err != nil {
		if errors.Is(err, io.EOF) {
			return base, nil
		}
		return Config{}, err
	}
	return fc.merge(base)
}
