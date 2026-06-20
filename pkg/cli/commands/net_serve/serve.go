package net_serve

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func serve() error {
	http.HandleFunc("/ip", handle(func(w http.ResponseWriter, r *http.Request) error {
		bytes, err := json.Marshal(struct {
			IP        string `json:"ip"`
			UserAgent string `json:"userAgent"`
		}{IP: r.RemoteAddr, UserAgent: r.UserAgent()})
		if err != nil {
			return errors.Wrap(err, "marshal resp")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes) //nolint: errcheck // no errors

		return nil
	}))

	println(color.GreenString("Serving on http://0.0.0.0:8000"))

	err := http.ListenAndServe(":8000", nil) //nolint: gosec // don't care
	if err != nil {
		return errors.Wrap(err, "listen and serve")
	}

	return nil
}

func Run(_ context.Context, _ *cli.Command) error {
	return serve()
}
