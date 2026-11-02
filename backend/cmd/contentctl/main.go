package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"knowledgegraph/backend/internal/content"
)

func main() {
	root := flag.String("root", "./var/documents", "document directory")
	database := flag.String("db", "./var/content.db", "SQLite database")
	legacy := flag.String("migrate-legacy", "", "legacy Markdown directory to migrate as private notes")
	seed := flag.String("seed", "", "explicit public seed directory")
	report := flag.String("report", "", "migration report path")
	updateFile := flag.String("update-file", "", "save a reviewed canonical Markdown file")
	expectedVersion := flag.String("version", "", "required working revision for --update-file")
	publish := flag.Bool("publish", false, "publish the saved --update-file revision")
	publishAll := flag.Bool("publish-all", false, "explicitly enable and publish every active document in all interface languages")
	flag.Parse()
	store, err := content.Open(*root, *database)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	if *updateFile != "" {
		if *expectedVersion == "" {
			log.Fatal("--update-file requires --version")
		}
		raw, err := os.ReadFile(*updateFile)
		if err != nil {
			log.Fatal(err)
		}
		d, err := content.Parse(raw)
		if err != nil {
			log.Fatal(err)
		}
		current, err := store.Get(d.ID, d.Locale)
		if err != nil {
			log.Fatal(err)
		}
		if current.Version != *expectedVersion {
			log.Fatal("VERSION_CONFLICT")
		}
		d.SourceRevision = current.SourceRevision
		saved, err := store.Write(ctx, d.ID, content.WriteInput{Title: d.Title, Summary: d.Summary, Slug: d.Slug, Kind: d.Kind, Category: d.Category, Tags: d.Tags, Content: d.Content, Locale: d.Locale, Version: *expectedVersion, Channels: &d.Channels, SourceLocale: d.SourceLocale, SourceRevision: d.SourceRevision})
		if err != nil {
			log.Fatal(err)
		}
		if *publish {
			if _, err = store.Publish(ctx, saved.ID, saved.Locale, saved.PublicationVersion); err != nil {
				log.Fatal(err)
			}
		}
		log.Print("document revision saved")
	}
	if *legacy != "" {
		entries, err := store.MigrateLegacy(ctx, *legacy)
		if err != nil {
			log.Fatal(err)
		}
		if *report != "" {
			raw, _ := json.MarshalIndent(entries, "", "  ")
			if err = os.MkdirAll(filepath.Dir(*report), 0700); err != nil {
				log.Fatal(err)
			}
			if err = os.WriteFile(*report, raw, 0600); err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("migrated %d private documents; original files unchanged", len(entries))
	}
	if *seed != "" {
		if err = store.Seed(ctx, *seed); err != nil {
			log.Fatal(err)
		}
		log.Print("seed content loaded; existing documents preserved")
	}
	if *publishAll {
		results, err := store.PublishAll(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if *report != "" {
			raw, _ := json.MarshalIndent(results, "", "  ")
			if err = os.MkdirAll(filepath.Dir(*report), 0700); err != nil {
				log.Fatal(err)
			}
			if err = os.WriteFile(*report, raw, 0600); err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("published %d documents across all interface languages", len(results))
	}
}
