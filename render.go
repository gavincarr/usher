/*
usher is a tiny personal url shortener.

Render (render.com) is a newish PAAS startup that offers free
hosting for static websites, with built-in HTTPS and CDN support,
as well as pretty interesting paid application hosting. For
usher, it's much easier to configure than Amazon S3.

This file contains functions for publishing the usher database
mappings as an infrastructure config `render.yaml` file. That
file can then be pushed via git for render.com to pick up i.e.
the Render publishing sequence is:

    usher push
	cd $(usher root)
	git push
*/

package usher

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

const configName = "render.yaml"

type Route struct {
	Type        string `yaml:"type"`
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
}

type Service struct {
	Type   string  `yaml:"type"`
	Name   string  `yaml:"name"`
	Env    string  `yaml:"env"`
	Routes []Route `yaml:"routes"`
}

type Config struct {
	Services []Service `yaml:"services"`
}

// pushRender publishes our usher database mappings as a
// infrastructure config `render.yaml` file for render.com.
// See https://render.com/docs/yaml-spec for the spec.
func (db *DB) pushRender() error {
	configfile := filepath.Join(db.Root, configName)

	// Check timestamps on database vs. configfile
	// This is an optimisation path, so we ignore errors
	statCF, err := os.Stat(configfile)
	if err == nil {
		statDB, err := os.Stat(db.DBPath)
		if err == nil {
			// If configfile is newer than database we can noop
			if statCF.ModTime() > statDB.ModTime() {
				return nil
			}
		}
	}

	// Read all mappings
	mappings, err := db.readDB()
	if err != nil {
		return err
	}

	// Assemble config
	config := Config{Services: make([]Service, 1)}
	service := Service{}
	service.Type = "web"
	service.Name = db.Domain
	service.Env = "static"
	service.Routes = make([]Route, len(mappings))
	config.Services[0] = service

	// Create code-url mappings
	i := 0
	for code, url := range mappings {
		service.Routes[i].Type = "redirect"
		service.Routes[i].Source = "/" + code
		service.Routes[i].Destination = url
		i++
	}

	// Output
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	tmpfile := configfile + ".tmp"
	err = ioutil.WriteFile(tmpfile, data, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(tmpfile, configfile)
	if err != nil {
		return err
	}

	return nil
}
