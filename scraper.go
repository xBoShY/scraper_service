package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Result struct {
	URL string
	StatusCode int
}

type Scraper struct { 
	Wait func() (err error) // always returns a non-nil error. After Close(), the returned error is ErrServerClosed.
	Close func() // Close the Scraper
	GetListen func() (addr string) // Returns the listening address
}

// maxWorkers limits the number of concurrent requests
func (m *Metrics) maxWorkers(h http.Handler, n uint8, timeout uint64) http.Handler {
	// use a synchronous channel of size n to limit the number of parallel requests by n
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// a request will either register (by sending to the channel) or timeout
		var timedout bool = false
		t1 := time.Now()
		select {
		case sema <- struct{}{}:
			// unlock the worker when processing completes
			defer func() { <-sema }()
		case <- time.After(time.Millisecond * time.Duration(timeout)):
			w.WriteHeader(http.StatusRequestTimeout)
			m.IncService(http.StatusRequestTimeout)
			timedout = true
		}

		// Calculate the wait time for an available worker
		t2 := time.Now()
		tWait := float64(int64(t2.Sub(t1) / time.Millisecond))
		if tWait > 0 {
			m.AddWorkerWait(tWait)
		}

		if timedout {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// requestHandler has all the request processing logic
func requestHandler(w http.ResponseWriter, r *http.Request) (*Result, int) {
	type RequestParams struct {
		URL string
	}

	var (
		data []byte // holds inbound data to be parsed and used
		err error
		reqParams RequestParams
	)

	// Implemented HTTP methods
	switch r.Method {
		case "POST":
			data, err = ioutil.ReadAll(r.Body)
		
		default:
			return nil, http.StatusNotImplemented
	}

	// Unexpected error reading the request data
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	// Parsing JSON to reqParams
	json.Unmarshal(data, &reqParams)

	// System's certificate pool might be outdated, certificate validation is not required
	tls := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{}
	tr.TLSClientConfig = tls
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	
	resp, err := client.Get(reqParams.URL)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	// Get method executed with success
	res := new(Result)
	res.URL = reqParams.URL
	res.StatusCode = resp.StatusCode
	
	return res, http.StatusOK
}

// requestHandlerWrapper registers the metrics logic
func (m *Metrics) requestHandlerWrapper (w http.ResponseWriter, r *http.Request) {
	result, response := requestHandler(w, r)

	w.WriteHeader(response)
	if result != nil {
		m.IncScrapes(result.URL, result.StatusCode)
	}
	m.IncService(response)
}

// NewScraper generates a Scraper, which will serve the requests on address <listen>
func (m *Metrics) NewScraper(endpoint string, listen string, workers uint8, timeout uint64) (*Scraper, error) {
	handler := http.HandlerFunc(m.requestHandlerWrapper)

	httpListener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle(endpoint, m.maxWorkers(handler, workers, timeout))
	httpSrv := &http.Server {
		Handler: mux,
	}

	//This channel will receive the error generated by http.Server.Serve
	//Its useful for the blocking funcion Scraper.Wait()
	ch := make(chan error)
	go func() {
		err := httpSrv.Serve(httpListener)
		ch <- err
	}()

	s := new(Scraper)
	s.Wait = func() (error) {
		return <-ch
	}

	s.Close = func() {
		httpSrv.Shutdown(context.Background())
		httpListener.Close()
	}

	s.GetListen = func() (string) {
		return httpListener.Addr().String()
	}

	return s, nil
}
