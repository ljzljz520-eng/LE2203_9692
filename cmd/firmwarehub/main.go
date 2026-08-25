package main

import (
	"flag"
	"fmt"
	"os"

	"firmwarehub/internal/archive"
	"firmwarehub/internal/catalog"
	"firmwarehub/internal/flow020"
	"firmwarehub/internal/importer"
	"firmwarehub/internal/review"
	"firmwarehub/internal/store"
)

func main() {
	dbPath := flag.String("db", "firmwarehub.db", "path to the bbolt database")
	flag.Parse()
	s, err := store.Open(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer s.Close()
	catalogService := catalog.NewService(s)
	reviewService := review.NewService(catalogService, review.DefaultPolicy())
	archiveService := archive.NewService(catalogService)
	application := flow020.NewApplication(catalogService, reviewService, archiveService, importer.NewProcessor(catalogService, s), archive.NewReporter(s))
	_ = application
	fmt.Printf("firmwarehub ready at %s\n", s.Path())
}
