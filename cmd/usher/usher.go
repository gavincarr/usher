package main

import (
	"fmt"
	"log"

	"github.com/alecthomas/kong"
	"github.com/gavincarr/usher"
)

var CLI struct {
	Init struct {
		Domain string `arg name:"domain" help:"Domain to be used for new database."`
	} `cmd help:"Initialise new usher database for domain."`

	Ls struct {
		Glob string `arg optional name:"glob" help:"Code glob of mappings to list."`
	} `cmd help:"List current mappings."`

	Add struct {
		Url  string `arg name:"url" help:"Url to redirect to."`
		Code string `arg optional name:"code" help:"Code to be used for mapping."`
	} `cmd help:"Add a new mapping to the usher database."`

	Rm struct {
		Code string `arg name:"code" help:"Code of mapping to remove from the database."`
	} `cmd help:"Remove a mapping from the usher database."`
}

func main() {
	log.SetFlags(0)
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {

	case "init <domain>":
		db, err := usher.NewDB(CLI.Init.Domain)
		if err != nil {
			log.Fatal(err)
		}
		created, err := db.Init()
		if err != nil {
			log.Fatal(err)
		}
		if created {
			fmt.Printf("Created new database %q\n", db.Filepath)
		} else {
			fmt.Printf("Database %q already exists\n", db.Filepath)
		}

	case "ls":
		db, err := usher.NewDB("")
		if err != nil {
			log.Fatal(err)
		}
		entries, err := db.List(CLI.Ls.Glob)
		if err != nil {
			log.Fatal(err)
		}
		for _, e := range entries {
			fmt.Printf("%-12s %s\n", e.Code, e.Url)
		}

	case "add <url> <code>":
		db, err := usher.NewDB("")
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Add(CLI.Add.Url, CLI.Add.Code)
		if err != nil {
			if err == usher.ErrCodeExists {
				log.Fatalf("Error: code %q already exists in usher database\n", CLI.Add.Code)
			} else {
				log.Fatal(err)
			}
		}

	case "add <url>":
		db, err := usher.NewDB("")
		if err != nil {
			log.Fatal(err)
		}
		code, err := db.Add(CLI.Add.Url, "")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Added mapping with code %q\n", code)

	case "rm <code>":
		db, err := usher.NewDB("")
		if err != nil {
			log.Fatal(err)
		}
		err = db.Remove(CLI.Rm.Code)
		if err != nil {
			if err == usher.ErrNotFound {
				log.Fatalf("Error: code %q not found in usher database\n", CLI.Rm.Code)
			} else {
				log.Fatal(err)
			}
		}

	default:
		log.Fatal(ctx.Command())
	}
}
