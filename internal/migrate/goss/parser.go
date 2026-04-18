package goss

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// GossFile represents a parsed Goss YAML configuration.
type GossFile struct {
	Package     map[string]GossAttrs `yaml:"package,omitempty"`
	Service     map[string]GossAttrs `yaml:"service,omitempty"`
	Process     map[string]GossAttrs `yaml:"process,omitempty"`
	Port        map[string]GossAttrs `yaml:"port,omitempty"`
	Command     map[string]GossAttrs `yaml:"command,omitempty"`
	File        map[string]GossAttrs `yaml:"file,omitempty"`
	User        map[string]GossAttrs `yaml:"user,omitempty"`
	Group       map[string]GossAttrs `yaml:"group,omitempty"`
	HTTP        map[string]GossAttrs `yaml:"http,omitempty"`
	DNS         map[string]GossAttrs `yaml:"dns,omitempty"`
	Addr        map[string]GossAttrs `yaml:"addr,omitempty"`
	Interface   map[string]GossAttrs `yaml:"interface,omitempty"`
	Mount       map[string]GossAttrs `yaml:"mount,omitempty"`
	KernelParam map[string]GossAttrs `yaml:"kernel-param,omitempty"`
	Gossfile    map[string]GossAttrs `yaml:"gossfile,omitempty"`
}

// GossAttrs holds resource attributes as a flexible map.
type GossAttrs map[string]interface{}

// Parse parses raw YAML bytes into a GossFile.
func Parse(data []byte) (*GossFile, error) {
	var gf GossFile
	if err := yaml.Unmarshal(data, &gf); err != nil {
		return nil, fmt.Errorf("parsing goss yaml: %w", err)
	}
	return &gf, nil
}
