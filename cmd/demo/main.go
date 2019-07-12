package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("index")
		_, _ = fmt.Fprintf(w, "Use /pass or /fail")
	})

	http.HandleFunc("/pass", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("pass")
		os.Exit(0)
	})

	http.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("fail")
		os.Exit(1)
	})

	fmt.Println("Listening on :8080")
	log.Println(http.ListenAndServe(":8080", nil))
}
