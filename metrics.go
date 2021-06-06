package main

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

// Metrics contains all the functions to manipulate the metrics
type Metrics struct {
	GetService func(status int) (value float64, err error) // Returns the current count of http_requests_total{status}
	IncService func(status int) // Increments http_requests_total{status}

	GetScrapes func(url string, status int) (value float64, err error) // Returns the current count of http_get{url,status}
	IncScrapes func(url string, status int) // Increments http_get{status}

	AddWorkerWait func(wait float64) // Adds wait time to the wait_available_worker{}
	GetCountWorkerWaits func() (value uint64, err error) // Returns the count of Adds to wait_available_worker{}
	GetSumWorkerWaits func() (value float64, err error) // Returns the sum of Adds to wait_available_worker{}

	Close func() // Closes the Metrics
}

// helper function
func toString(code int) string {
	return strconv.Itoa(code)
}

// NewMetrics generates Metrics and its functions implementation using prometheus
func NewMetrics(endpoint string, listen string) (*Metrics, error) {
	//available metrics
	scrapes := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_get",
			Help: "How many HTTP GET scraped, partitioned by url and response code.",
		},
		[]string{"code","url"},
	)

	service := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests received, partitioned response code.",
		},
		[]string{"code"},
	)

	workerWait := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "wait_available_worker",
			Help:    "Time waiting for an available worker. In milliseconds.",
			Buckets: []float64{10, 100, 1000, 2000, 5000, 10000},
		},
	)

	// Creates a prometheus registry and register the previous metrics 
	reg := prometheus.NewRegistry()
	reg.MustRegister(scrapes)
	reg.MustRegister(service)
	reg.MustRegister(workerWait)

	httpListener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}

	// Creates a handler for prometheus registry
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	mux := http.NewServeMux()
	mux.Handle(endpoint, handler)
	httpSrv := &http.Server {
		Handler: mux,
	}

	go httpSrv.Serve(httpListener)

	metrics := new(Metrics)

	metrics.GetService = func(c int) (float64, error) {
		var m = &dto.Metric{}
		err := service.WithLabelValues(toString(c)).Write(m)
		if (err != nil) {
			return 0, nil
		}
		
		return m.Counter.GetValue(), nil
	}

	metrics.IncService = func(c int) {
		service.WithLabelValues(toString(c)).Inc()
	}

	metrics.GetScrapes = func(url string, c int) (float64, error) {
		var m = &dto.Metric{}
		err := scrapes.WithLabelValues(url, toString(c)).Write(m)
		if (err != nil) {
			return 0, nil
		}
		
		return m.Counter.GetValue(), nil
	}

	metrics.IncScrapes = func(url string, c int) {
		scrapes.WithLabelValues(url, toString(c)).Inc()
	}

	metrics.AddWorkerWait = func(w float64) {
		workerWait.Observe(w)
	}

	metrics.GetCountWorkerWaits = func() (uint64, error) {
		var m = &dto.Metric{}
		err := workerWait.Write(m)
		if (err != nil) {
			return 0, nil
		}
		
		return m.Histogram.GetSampleCount(), nil
	}

	metrics.GetSumWorkerWaits = func() (float64, error) {
		var m = &dto.Metric{}
		err := workerWait.Write(m)
		if (err != nil) {
			return 0, nil
		}
		
		return m.Histogram.GetSampleSum(), nil
	}

	metrics.Close = func() {
		httpSrv.Shutdown(context.Background())
		httpListener.Close()
		scrapes.Reset()
		service.Reset()
		reg.Unregister(scrapes)
		reg.Unregister(service)
		reg.Unregister(workerWait)
	}

	return metrics, nil
}
