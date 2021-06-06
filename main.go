package main

import (
	"log"

	"github.com/cosiner/flag"
)

// Arguments defines the available app arguments
type Arguments struct {
	Listen	      string	`names:"--listen" env:"SCRAPER_Listen" usage:"Service listen address." default:":8080"`
	Workers	      uint8		`names:"--workers" env:"SCRAPER_Workers" usage:"Number of serving workers." default:"2"`
	Timeout       uint64  `names:"--timeout" env:"SCRAPER_Timeout" usage:"Maximum time (in milliseconds) to wait for a worker." default:"1000"`
	MetricsListen string	`names:"--metrics" env:"SCRAPER_MetricsListen" usage:"Metrics listen address." default:":9095"`
}

// Service contains the Metrics logic and the Scraper logic
type Service struct {
	metrics *Metrics
	scraper *Scraper
}

// StartService generates and start a Service
func StartService(args Arguments) (*Service) {
	m, err := NewMetrics("/metrics", args.MetricsListen)
	if err != nil {
		log.Fatal(err)
	}

	s, err := m.NewScraper("/", args.Listen, args.Workers, args.Timeout)
	if err != nil {
		log.Fatal(err)
	}

	svc := new(Service)
	svc.metrics = m
	svc.scraper = s

	return svc
}

func main() {
	var args Arguments
  flag.Commandline.ParseStruct(&args)

	service := StartService(args)

	defer service.metrics.Close()
	log.Fatal(service.scraper.Wait())
}
