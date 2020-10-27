/*
usher is a tiny personal url shortener.

Render (render.com) is a newish PAAS startup that offers free
hosting for static websites, with built-in HTTPS and CDN support,
as well as nice paid application hosting. For usher, it's much
easier to configure than Amazon S3.

This file contains functions for publishing the usher database
mappings as an infrastructure config `render.yaml` file. That
file can then be pushed via git for render.com to pick up i.e.
the Render publishing sequence is:

    usher push
	cd $(usher root)
	git push

See `Render.md` for more details on setting up on render.com.
*/

package usher

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

const configName = "render.yaml"
const buildPath = "./build"

type Route struct {
	Type        string `yaml:"type"`
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
}

type Service struct {
	Type         string  `yaml:"type"`
	Name         string  `yaml:"name"`
	Env          string  `yaml:"env"`
	BuildCommand string  `yaml:"buildCommand"`
	BuildPath    string  `yaml:"staticPublishPath"`
	Routes       []Route `yaml:"routes"`
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
			if statCF.ModTime().After(statDB.ModTime()) {
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
	service.BuildCommand = ""
	service.BuildPath = buildPath
	service.Routes = make([]Route, len(mappings))
	config.Services[0] = service

	// Extract codes and sort (or render.yaml routes are randomly ordered)
	codes := make([]string, len(mappings))
	i := 0
	for code := range mappings {
		codes[i] = code
		i++
	}
	sort.Strings(codes)

	// Create code-url mappings
	i = 0
	for _, code := range codes {
		service.Routes[i].Type = "redirect"
		// Handle `indexCode` specially
		if code == indexCode {
			service.Routes[i].Source = "/"
		} else {
			service.Routes[i].Source = "/" + code
		}
		service.Routes[i].Destination = mappings[code]
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

	// Render seems to require our BuildPath to actually exist, so add `buildPath/.gitignore` if missing
	buildDir := filepath.Join(db.Root, buildPath)
	_, err = os.Stat(buildDir)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(buildDir, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	gifile := filepath.Join(buildDir, ".gitignore")
	gitignore := `*
!.gitignore
`
	_, err = os.Stat(gifile)
	if err != nil && os.IsNotExist(err) {
		err = ioutil.WriteFile(gifile, []byte(gitignore), 0644)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
