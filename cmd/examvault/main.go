package main

import (
	"fmt"
	"os"

	"examvault/internal/api"
	"examvault/internal/flow"
	"examvault/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("examvault: register, review, confirm, update, publish, archive, search, import, refresh")
		return
	}
	path := "examvault.db"
	if value := os.Getenv("EXAMVAULT_DB"); value != "" {
		path = value
	}
	database, err := store.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer database.Close()
	service := flow.NewService(database, "examvault-deterministic-key", 1, "cli")
	output, err := (api.Runner{Service: service}).Run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(output)
}
