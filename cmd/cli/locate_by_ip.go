package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"net/http"
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

func (r *CLI) commandLocateByIP(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() == 0 {
		return errors.New("need at least one IP address or domain")
	}

	ipOrDomain := command.Args().First()
	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://ip-api.com/json/%s", ipOrDomain))
	if err != nil {
		return errors.Wrap(err, "failed to get resp")
	}

	var domainLocation DomainLocationResp
	err = json.NewDecoder(resp.Body).Decode(&domainLocation)
	if err != nil {
		return errors.Wrap(err, "failed to decode resp")
	}
	if domainLocation.Status != "success" {
		return errors.Errorf("failed to locate IP, %s, query: %s", domainLocation.Message, domainLocation.Query)
	}

	fmt.Printf("Address: %s %s %s\n", color.GreenString(domainLocation.Country), domainLocation.RegionName, domainLocation.City)
	fmt.Printf("Coordinates: %s (https://yandex.ru/maps/?ll=%f%%2C%f&z=16)\n", color.YellowString(fmt.Sprintf("%f,%f", domainLocation.Lat, domainLocation.Lon)), domainLocation.Lon, domainLocation.Lat)
	fmt.Printf("Organization: %s, %s, %s\n", domainLocation.Isp, color.BlueString(domainLocation.Org), domainLocation.As)

	return nil
}
