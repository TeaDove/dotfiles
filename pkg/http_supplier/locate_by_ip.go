package http_supplier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
)

type DomainLocationResp struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Query       string  `json:"query"`
}

func (r *Supplier) LocateByIP(ctx context.Context, ip string) (DomainLocationResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://ip-api.com/json/%s", ip), nil)
	if err != nil {
		return DomainLocationResp{}, errors.Wrap(err, "build request to get ip")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return DomainLocationResp{}, errors.Wrap(err, "get resp")
	}

	var domainLocation DomainLocationResp

	err = json.NewDecoder(resp.Body).Decode(&domainLocation)
	if err != nil {
		return DomainLocationResp{}, errors.Wrap(err, "decode resp")
	}

	if domainLocation.Status != "success" {
		return DomainLocationResp{}, errors.Newf(
			"failed to locate IP, %s, query: %s",
			domainLocation.Message,
			domainLocation.Query,
		)
	}

	if domainLocation.City == domainLocation.RegionName {
		domainLocation.City = ""
	}

	return domainLocation, nil
}
