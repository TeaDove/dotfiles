package net_serve

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
)

func handle(fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": %s}`, err.Error()) //nolint: gosec // no taint
			fmt.Printf("%s: %s\n", color.RedString("Unexpected error in http handler"), err.Error())
		}
	}
}
