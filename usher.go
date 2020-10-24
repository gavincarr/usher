/*
usher is a tiny personal url shortener.

This library provides the maintenance functions for our simple
database of code => url mappings.
*/

package usher

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

var (
	ErrNotFound = errors.New("Not Found")
)

type DB struct {
	Root     string // full path to usher root directory containing databases
	Domain   string // fully-qualified domain whose mappings we want
	Filepath string // full database filepath to database for Domain
}

type Entry struct {
	Code string
	Url  string
}

// NewDB creates a DB struct with members derived from parameters,
// the environment, or defaults (in that order). It does no checking
// that the values produced are sane or exist on the filesystem.
func NewDB(domain string) (*DB, error) {
	// Get root
	root := os.Getenv("USHER_ROOT")
	if root == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(configDir, "usher")
	}

	// Get domain
	if domain == "" {
		domain = os.Getenv("USHER_DOMAIN")
	}
	if domain == "" {
		return nil, errors.New("Domain not passed as parameter or set in env USHER_DOMAIN")
	}

	// Set filepath
	path := filepath.Join(root, domain+".yml")

	return &DB{Root: root, Domain: domain, Filepath: path}, nil
}

// Init checks whether an usher root exists, creating it, if not,
// and then checks whether an usher database exists for domain,
// creating it if not.
func (db *DB) Init() (created bool, err error) {
	// Ensure root exists
	err = os.MkdirAll(db.Root, 0755)
	if err != nil {
		return false, err
	}

	// Ensure database exists
	_, err = os.Stat(db.Filepath)
	if err == nil {
		return false, nil // exists
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err // unexpected error
	}

	// Database does not exist - create
	fh, err := os.Create(db.Filepath)
	fh.Close()
	if err != nil {
		return false, err
	}
	return true, nil
}

// List returns the set of database entries whose code matches glob
func (db *DB) List(glob string) ([]Entry, error) {
	// FIXME: first-pass - ignore glob
	mappings, err := db.readfile()
	if err != nil {
		return nil, err
	}

	// Extract codes and sort
	codes := make([]string, len(mappings))
	i := 0
	for code := range mappings {
		codes[i] = code
		i++
	}
	sort.Strings(codes)

	// Compile entries
	var entries = make([]Entry, len(mappings))
	i = 0
	for _, code := range codes {
		entries[i] = Entry{Code: code, Url: mappings[code]}
		i++
	}

	return entries, nil
}

func (db *DB) Add(url, code string) error {
	return nil
}

// Remove the mapping with code from the database
// Returns ErrNotFound if code does not exist in the database
func (db *DB) Remove(code string) error {
	mappings, err := db.readfile()
	if err != nil {
		return err
	}

	_, exists := mappings[code]
	if !exists {
		return ErrNotFound
	}

	delete(mappings, code)

	err = db.writefile(mappings)
	if err != nil {
		return err
	}

	return nil
}

// readfile is a utility function to read all mappings from db.Filepath
// and return as a go map
func (db *DB) readfile() (map[string]string, error) {
	data, err := ioutil.ReadFile(db.Filepath)
	if err != nil {
		return nil, err
	}

	var mappings map[string]string
	err = yaml.Unmarshal(data, &mappings)
	if err != nil {
		return nil, err
	}

	return mappings, nil
}

// writefile is a utility function to write mappings to db.Filepath
func (db *DB) writefile(mappings map[string]string) error {
	data, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}

	tmpfile := db.Filepath + ".tmp"
	err = ioutil.WriteFile(tmpfile, data, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(tmpfile, db.Filepath)
	if err != nil {
		return err
	}

	return nil
}
