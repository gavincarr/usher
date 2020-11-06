// Usher unit tests

package usher

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/udhos/equalfile"
)

const (
	testDir          = "testdata"
	testRootExisting = "testdata/root"
	testRootNew      = "testdata/root2"
	testGolden       = "testdata/golden"
	domain           = "example.me"
	dbfile           = "example.me.yml"
)

// TestBasic runs integration tests from an existing root directory
func TestBasic(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	doSetupExisting(t, cwd)
	cmp := equalfile.New(nil, equalfile.Options{})

	db := testNewDBDomain(t, cwd, testRootExisting, domain)
	testInit(t, db, cwd, cmp)
	testAdd(t, db, cmp, cwd, "test1", "https://example.com/test1", "add1.yml")
	db2 := testNewDBInferDomain(t, cwd, domain)
	testAdd(t, db2, cmp, cwd, "test2", "https://example.com/test2", "add2.yml")
	testUpdate(t, db, cmp, cwd, "test2", "https://example.com/test3", "update1.yml")
	testAdd(t, db, cmp, cwd, "test4", "https://example.com/test4", "add3.yml")
	testPushRender(t, db, cmp, cwd)
	code := testAddRandom(t, db, cmp, "https://example.com/test5")
	testList(t, db, []string{"test1", "test2", "test4", code})
	testRemove(t, db, cmp, cwd, code, "add3.yml")
	testRemove(t, db, cmp, cwd, "test1", "remove1.yml")
	testRemove(t, db, cmp, cwd, "test4", "remove2.yml")
	testRemove(t, db, cmp, cwd, "test2", "empty.yml")

	// Reset cwd
	err = os.Chdir(cwd)
	if err != nil {
		t.Fatal(err)
	}
}

// TestNewRoot runs integration tests from outside a new root directory
func TestNewRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	doSetupNew(t, cwd)
	cmp := equalfile.New(nil, equalfile.Options{})

	db := testNewDBDomain(t, cwd, testRootNew, domain)
	testInit(t, db, cwd, cmp)
	testAdd(t, db, cmp, cwd, "test1", "https://example.com/test1", "add1.yml")
	testAdd(t, db, cmp, cwd, "test2", "https://example.com/test2", "add2.yml")
	testUpdate(t, db, cmp, cwd, "test2", "https://example.com/test3", "update1.yml")
	testAdd(t, db, cmp, cwd, "test4", "https://example.com/test4", "add3.yml")
	code := testAddRandom(t, db, cmp, "https://example.com/test5")
	testList(t, db, []string{"test1", "test2", "test4", code})
	testRemove(t, db, cmp, cwd, code, "add3.yml")
}

func testNewDBDomain(t *testing.T, cwd, root, domain string) *DB {
	db, err := NewDB(domain)
	if err != nil {
		t.Fatal(err)
	}
	if db.Root != filepath.Join(cwd, root) {
		t.Fatalf("NewDB root %q differs from expected %q",
			db.Root, filepath.Join(cwd, root))
	}
	if db.Domain != domain {
		t.Fatalf("db domain %q differs from expected %q", db.Domain, domain)
	}
	return db
}

func testNewDBInferDomain(t *testing.T, cwd, domain string) *DB {
	db, err := NewDB("")
	if err != nil {
		t.Fatal(err)
	}
	if db.Domain != domain {
		t.Errorf("db2 domain %q differs from expected %q", db.Domain, domain)
	}
	return db
}

func testInit(t *testing.T, db *DB, cwd string, cmp *equalfile.Cmp) {
	created, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, created, "new database created by Init()")

	equal, err := cmp.CompareFile(
		filepath.Join(db.Root, configfile),
		filepath.Join(cwd, testGolden, configfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Init() config %q differs from expected", configfile)
	}
}

func testAdd(t *testing.T, db *DB, cmp *equalfile.Cmp, cwd, code, url, goldenfile string) {
	_, err := db.Add(url, code)
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(
		filepath.Join(db.Root, dbfile),
		filepath.Join(cwd, testGolden, goldenfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Add() db %q differs from expected %q", dbfile, goldenfile)
	}
}

func testAddRandom(t *testing.T, db *DB, cmp *equalfile.Cmp, url string) string {
	code, err := db.Add(url, "")
	if err != nil {
		t.Fatal(err)
	}

	return code
}

func testUpdate(t *testing.T, db *DB, cmp *equalfile.Cmp, cwd, code, url, goldenfile string) {
	err := db.Update(url, code)
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(
		filepath.Join(db.Root, dbfile),
		filepath.Join(cwd, testGolden, goldenfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Update() db %q differs from expected %q", dbfile, goldenfile)
	}
}

func testList(t *testing.T, db *DB, codes []string) {
	entries, err := db.List("")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != len(codes) {
		t.Errorf("List() returned %d entries, expected %d\n", len(entries), len(codes))
	}

	// All entries should exist in codes
	codeMap := make(map[string]bool)
	for _, code := range codes {
		codeMap[code] = true
	}
	for _, entry := range entries {
		if _, exists := codeMap[entry.Code]; !exists {
			t.Errorf("List entry %q not found in expected codes: %v\n", entry.Code, codes)
		}
	}
}

func testRemove(t *testing.T, db *DB, cmp *equalfile.Cmp, cwd, code, goldenfile string) {
	err := db.Remove(code)
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(
		filepath.Join(db.Root, dbfile),
		filepath.Join(cwd, testGolden, goldenfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Remove() db %q differs from expected %q", dbfile, goldenfile)
	}
}

func testPushRender(t *testing.T, db *DB, cmp *equalfile.Cmp, cwd string) {
	outfile := "render.yaml"

	// Remove any existing outfile
	_, err := os.Stat(outfile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	} else if err == nil {
		err = os.Remove(outfile)
		if err != nil {
			t.Fatal(err)
		}
	}

	// If we use the existing configfile, we should get an ErrPushTypeUnconfigured error
	pushErr := db.Push()
	if pushErr == nil {
		t.Error("Push() with unconfigured config did not return an error!")
	}
	if pushErr != ErrPushTypeUnconfigured {
		t.Errorf("Unexpected error from Push() with unconfigured config: %s\n", pushErr)
	}

	// If we replace configfile with a bogus type, we should get an ErrPushTypeBad error
	err = db.writeConfigString(db.Domain + `:
  type: bogus
`)
	if err != nil {
		t.Fatal(err)
	}
	pushErr = db.Push()
	if pushErr == nil {
		t.Error("Push() with bogus type config did not return an error!")
	}
	if !errors.Is(pushErr, ErrPushTypeBad) {
		t.Fatalf("Unexpected error from Push() with bogus type config: %s\n", pushErr)
	}

	// Replace configfile with render version
	err = db.writeConfigString(db.Domain + `:
  type: render
`)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Push()
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(
		filepath.Join(db.Root, outfile),
		filepath.Join(cwd, testGolden, outfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Push() output file %q differs from expected", outfile)
	}
}

// doSetupExisting is a utility function to prep an existing root directory for testing
func doSetupExisting(t *testing.T, cwd string) {
	// Unset usher environment variables
	err := os.Unsetenv("USHER_ROOT")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Unsetenv("USHER_DOMAIN")
	if err != nil {
		t.Fatal(err)
	}

	// Chdir to testRootExisting for this test (precondition for Init() to work)
	err = os.Chdir(filepath.Join(cwd, testRootExisting))
	if err != nil {
		t.Fatal(err)
	}

	// Remove any existing db
	_, err = os.Stat(dbfile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	} else if err == nil {
		err = os.Remove(dbfile)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Truncate any existing config (not remove, so can still infer the root)
	fh, err := os.Create(configfile)
	fh.Close()
}

// doSetupNew is a utility function to prep a new root directory for testing,
// using USHER_ROOT
func doSetupNew(t *testing.T, cwd string) {
	// Set/unset usher environment variables
	root, err := filepath.Abs(testRootNew)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("USHER_ROOT", root)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Unsetenv("USHER_DOMAIN")
	if err != nil {
		t.Fatal(err)
	}

	// Check that the testDir directory exists, to ensure Cwd is correct
	_, err = os.Stat(filepath.Join(cwd, testDir))
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("test directory %q does not exist!", filepath.Join(cwd, testDir))
	} else if err != nil {
		t.Fatal(err)
	}

	// Remove any testRootExisting directory and contents (succeeds if missing)
	err = os.RemoveAll(filepath.Join(cwd, testRootNew))
	if err != nil {
		t.Fatal(err)
	}
}
