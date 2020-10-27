/*
usher is a tiny personal url shortener.

This library provides the maintenance functions for our simple
database of code => url mappings (a yaml file in
filepath.join(os.UserConfigDir(), "usher")).
*/

package usher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const configfile = "usher.yml"
const indexCode = "INDEX"

// Random Code generation constants
const minRandomCodeLen = 5
const maxRandomCodeLen = 8
const digits = "23456789"                // omit 0 and 1 as easily confused with o and l
const chars = "abcdefghijkmnpqrstuvwxyz" // omit o and l as easily confused with 0 and 1

var (
	ErrNotFound   = errors.New("not found")
	ErrCodeExists = errors.New("code already used")
	ErrNoChange   = errors.New("mapping unchanged")
)

type DB struct {
	Root       string // full path to usher root directory containing databases
	Domain     string // fully-qualified domain whose mappings we want
	DBPath     string // full path to database for Domain
	ConfigPath string // full path to usher config file
}

type Entry struct {
	Code string
	Url  string
}

type ConfigEntry struct {
	Type      string `yaml:"type"`
	AWSKey    string `yaml:"aws_key,omitempty"`
	AWSSecret string `yaml:"aws_secret,omitempty"`
	AWSRegion string `yaml:"aws_region,omitempty"`
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

	// Derive domain if not set - check for USHER_DOMAIN in environment
	if domain == "" {
		domain = os.Getenv("USHER_DOMAIN")
	}
	// Else infer the domain if only one database exists
	if domain == "" {
		matches, _ := filepath.Glob(filepath.Join(root, "*.*.yml"))
		if len(matches) == 1 {
			// Exactly one match - strip .yml suffix to get domain
			re := regexp.MustCompile(`.yml$`)
			domain = re.ReplaceAllLiteralString(filepath.Base(matches[0]), "")
		}
	}
	// Else give up with an error
	if domain == "" {
		return nil, errors.New("Domain not passed as parameter or set in env USHER_DOMAIN")
	}

	// Set DBPath
	dbpath := filepath.Join(root, domain+".yml")

	// Set ConfigPath
	configpath := filepath.Join(root, configfile)

	return &DB{Root: root, Domain: domain, DBPath: dbpath, ConfigPath: configpath}, nil
}

// Init checks and creates the following, if they don't exist:
// - an usher root directory
// - an usher database for the db.Domain
// - an entry in the user config file for db.Domain
func (db *DB) Init() (dbCreated bool, err error) {
	dbCreated = false

	// Ensure root exists
	err = os.MkdirAll(db.Root, 0755)
	if err != nil {
		return dbCreated, err
	}

	// Ensure database exists
	_, err = os.Stat(db.DBPath)
	if err == nil {
		return dbCreated, nil // exists
	}
	if err != nil && !os.IsNotExist(err) {
		return dbCreated, err // unexpected error
	}

	// Database does not exist - create
	fh, err := os.Create(db.DBPath)
	fh.Close()
	if err != nil {
		return dbCreated, err
	}
	dbCreated = true

	// Ensure configfile exists
	_, err = os.Stat(db.ConfigPath)
	if err == nil {
		_, err := db.readConfig()
		if err != nil {
			if err != ErrNotFound {
				return dbCreated, err
			}
		}
		err = db.appendConfigString(db.configPlaceholder())
		if err != nil {
			return dbCreated, err
		}
	} else {
		// Create a placeholder config file for domain
		err = db.writeConfigString(db.configPlaceholder())
		if err != nil {
			return dbCreated, err
		}
	}

	return dbCreated, nil
}

// List returns the set of database entries whose code matches glob
func (db *DB) List(glob string) ([]Entry, error) {
	// FIXME: first-pass - ignore glob
	mappings, err := db.readDB()
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

// Add a mapping for url and code to the database.
// If code is missing, a random code will be generated and returned.
func (db *DB) Add(url, code string) (string, error) {
	mappings, err := db.readDB()
	if err != nil {
		return "", err
	}

	if code == "" {
		code = randomCode(mappings)

	} else {
		// Check for parameter inversion
		reUrl := regexp.MustCompile(`^https?://`)
		if !reUrl.MatchString(url) && reUrl.MatchString(code) {
			url, code = code, url
		}

		// Check whether code is already used
		dburl, exists := mappings[code]
		if exists {
			if dburl == url {
				// Trying to re-add the same url is not an error, just a noop
				return code, nil
			}
			return code, ErrCodeExists
		}
	}

	mappings[code] = url
	err = db.writeDB(mappings)
	if err != nil {
		return code, err
	}

	return code, nil
}

// Update an existing mapping in the database, changing the URL.
func (db *DB) Update(url, code string) error {
	mappings, err := db.readDB()
	if err != nil {
		return err
	}

	// Check for parameter inversion
	reUrl := regexp.MustCompile(`^https?://`)
	if !reUrl.MatchString(url) && reUrl.MatchString(code) {
		url, code = code, url
	}

	// If code is missing, abort
	dburl, exists := mappings[code]
	if !exists {
		return ErrNotFound
	}

	// Trying to update to the same url is not an error, just a noop
	if dburl == url {
		return nil
	}

	mappings[code] = url
	err = db.writeDB(mappings)
	if err != nil {
		return err
	}

	return nil
}

// Remove the mapping with code from the database
// Returns ErrNotFound if code does not exist in the database
func (db *DB) Remove(code string) error {
	mappings, err := db.readDB()
	if err != nil {
		return err
	}

	_, exists := mappings[code]
	if !exists {
		return ErrNotFound
	}

	delete(mappings, code)

	err = db.writeDB(mappings)
	if err != nil {
		return err
	}

	return nil
}

// Push syncs all current mappings with the backend configured for db.Domain
// in db.ConfigPath
func (db *DB) Push() error {
	config, err := db.readConfig()
	if err != nil {
		return err
	}

	if config.Type == "" {
		return fmt.Errorf("no 'type' field found for %q in config %q\n",
			db.Domain, db.ConfigPath)
	}

	switch config.Type {
	case "s3":
		err = db.pushS3(config)
		if err != nil {
			return err
		}
	case "render":
		err = db.pushRender()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid type %q found for %q in config %q\n",
			config.Type, db.Domain, db.ConfigPath)
	}

	return nil
}

// readDB is a utility function to read all mappings from db.DBPath
// and return as a go map
func (db *DB) readDB() (map[string]string, error) {
	data, err := ioutil.ReadFile(db.DBPath)
	if err != nil {
		return nil, err
	}

	var mappings map[string]string
	err = yaml.Unmarshal(data, &mappings)
	if err != nil {
		return nil, err
	}

	if len(mappings) == 0 {
		mappings = make(map[string]string)
	}

	return mappings, nil
}

// writeDB is a utility function to write mappings (as yaml) to db.DBPath
func (db *DB) writeDB(mappings map[string]string) error {
	data, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}

	tmpfile := db.DBPath + ".tmp"
	err = ioutil.WriteFile(tmpfile, data, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(tmpfile, db.DBPath)
	if err != nil {
		return err
	}

	return nil
}

// readConfig is a utility function to read the config entry for
// db.Domain from db.ConfigPath file
func (db *DB) readConfig() (*ConfigEntry, error) {
	data, err := ioutil.ReadFile(db.ConfigPath)
	if err != nil {
		return nil, err
	}

	var entries map[string]ConfigEntry
	err = yaml.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	entry, exists := entries[db.Domain]
	if !exists {
		return nil, ErrNotFound
	}

	return &entry, nil
}

// writeConfigString is a utility function to write data to db.ConfigPath
func (db *DB) writeConfigString(data string) error {
	tmpfile := db.ConfigPath + ".tmp"
	err := ioutil.WriteFile(tmpfile, []byte(data), 0600)
	if err != nil {
		return err
	}
	err = os.Rename(tmpfile, db.ConfigPath)
	if err != nil {
		return err
	}

	return nil
}

// appendConfigString is a utility function to write data to db.ConfigPath
func (db *DB) appendConfigString(data string) error {
	config, err := ioutil.ReadFile(db.ConfigPath)
	if err != nil {
		return err
	}

	config = append(config, []byte(data)...)

	tmpfile := db.ConfigPath + ".tmp"
	err = ioutil.WriteFile(tmpfile, config, 0600)
	if err != nil {
		return err
	}
	err = os.Rename(tmpfile, db.ConfigPath)
	if err != nil {
		return err
	}

	return nil
}

// randomCode is a utility function to generate a random code
// and check that it doesn't exist in mappings.
// Random codes use the following pattern: 1 digit, then 4-7
// lowercase ascii characters. This usually allows them to be
// relatively easily distinguished from explicit codes, while
// still being easy to communicate orally.
func randomCode(mappings map[string]string) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	b.WriteByte(digits[rand.Intn(len(digits))])
	for i := 1; i < maxRandomCodeLen; i++ {
		b.WriteByte(chars[rand.Intn(len(chars))])
		// If long enough, check if exists in mappings, and return if not
		if i+1 >= minRandomCodeLen {
			s := b.String()
			if _, exists := mappings[s]; !exists {
				return s
			}
		}
	}
	// Failed to find an unused code? Just retry?
	return randomCode(mappings)
}

func (db *DB) configPlaceholder() string {
	return db.Domain + `:
# type: s3
# aws_key: foo
# aws_secret: bar
# aws_region: us-east-1
`
}
