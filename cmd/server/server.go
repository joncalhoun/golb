package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/joncalhoun/golb"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mux := http.NewServeMux()

	postTemplate := template.Must(template.ParseFiles("post.gohtml"))
	mux.HandleFunc("GET /posts/{slug}", golb.PostHandler(golb.FileReader{}, postTemplate))
	indexTemplate := template.Must(template.ParseFiles("index.gohtml"))
	mux.HandleFunc("GET /", golb.IndexHandler(golb.FileReader{}, indexTemplate))

	mux.HandleFunc("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))).ServeHTTP)

	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	return nil
}
