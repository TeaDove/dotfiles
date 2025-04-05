package http_supplier

import (
	"net/http"
	"time"
)

type Supplier struct {
	client *http.Client
}

func New() *Supplier {
	r := &Supplier{}
	r.client = &http.Client{Timeout: 10 * time.Second}

	return r
}
