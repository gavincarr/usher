// Usher unit tests

package usher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/udhos/equalfile"
)

const (
	testRoot   = "testdata/root"
	testGolden = "../golden"
	domain     = "example.me"
	dbfile     = "example.me.yml"
)

func TestBasic(t *testing.T) {
	doSetup(t)
	cmp := equalfile.New(nil, equalfile.Options{})

	db, err := NewDB(domain)
	if err != nil {
		t.Fatal(err)
	}
	testInit(t, db, cmp)
	testAdd(t, db, cmp, "test1", "https://example.com/test1", "add1.yml")
	testAdd(t, db, cmp, "test2", "https://example.com/test2", "add2.yml")
	testUpdate(t, db, cmp, "test2", "https://example.com/test3", "update1.yml")
	testAdd(t, db, cmp, "test4", "https://example.com/test4", "add3.yml")
	testPushRender(t, db, cmp)
	testRemove(t, db, cmp, "test1", "remove1.yml")
	testRemove(t, db, cmp, "test4", "remove2.yml")
	testRemove(t, db, cmp, "test2", "empty.yml")
}

func testInit(t *testing.T, db *DB, cmp *equalfile.Cmp) {
	created, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, created, true, "new database created by Init()")
	equal, err := cmp.CompareFile(
		filepath.Join(configfile),
		filepath.Join(testGolden, configfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Init() config %q differs from expected", configfile)
	}
}

func testAdd(t *testing.T, db *DB, cmp *equalfile.Cmp, code, url, goldenfile string) {
	_, err := db.Add(url, code)
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(dbfile, filepath.Join(testGolden, goldenfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Add() db %q differs from expected %q", dbfile, goldenfile)
	}
}

func testUpdate(t *testing.T, db *DB, cmp *equalfile.Cmp, code, url, goldenfile string) {
	err := db.Update(url, code)
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(dbfile, filepath.Join(testGolden, goldenfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Update() db %q differs from expected %q", dbfile, goldenfile)
	}
}

func testRemove(t *testing.T, db *DB, cmp *equalfile.Cmp, code, goldenfile string) {
	err := db.Remove(code)
	if err != nil {
		t.Fatal(err)
	}

	equal, err := cmp.CompareFile(dbfile, filepath.Join(testGolden, goldenfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Remove() db %q differs from expected %q", dbfile, goldenfile)
	}
}

func testPushRender(t *testing.T, db *DB, cmp *equalfile.Cmp) {
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

	equal, err := cmp.CompareFile(outfile, filepath.Join(testGolden, outfile))
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Errorf("post-Push() output file %q differs from expected", outfile)
	}
}

// setup is a utility function to set up our root directory for testing
func doSetup(t *testing.T) {
	err := os.Chdir(testRoot)
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

	// Truncate any existing config
	fh, err := os.Create(configfile)
	fh.Close()
}
