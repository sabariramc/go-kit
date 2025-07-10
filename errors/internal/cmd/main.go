package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sabariramc/go-kit/errors/internal"
)

func main() {
	mux := internal.NewServer()
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
