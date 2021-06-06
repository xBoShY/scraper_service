package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestServiceStatusOk(t *testing.T) {
	args := Arguments{
		Listen: ":0",
		Workers: 1,
		Timeout: 1000,
		MetricsListen: ":0",
	}

	url := "http://www.phaidra.ai"
	expectedServiceStatus := http.StatusOK
	expectedScrapeStatus := http.StatusOK

	svc := StartService(args)
	defer svc.metrics.Close()
	defer svc.scraper.Close()

	json := fmt.Sprintf("{\"url\": \"%s\"}", url)
	req := strings.NewReader(json)
	resp, err := http.Post("http://" + svc.scraper.GetListen(), "", req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != expectedServiceStatus {
		t.Fatalf(`Failed`)
	}

	serviceMetrics, err := svc.metrics.GetService(expectedServiceStatus)
	if err != nil || serviceMetrics != 1 {
		t.Fatalf(`Failed`)
	}

	scrapeMetrics, err := svc.metrics.GetScrapes(url, expectedScrapeStatus)
	if err != nil || scrapeMetrics != 1 {
		t.Fatalf(`Failed`)
	}
}

func TestServiceStatusTeapot(t *testing.T) {
	args := Arguments{
		Listen: ":0",
		Workers: 1,
		Timeout: 1000,
		MetricsListen: ":0",
	}

	url := "http://www.google.com/teapot"
	expectedServiceStatus := http.StatusOK
	expectedScrapeStatus := http.StatusTeapot

	svc := StartService(args)
	defer svc.metrics.Close()
	defer svc.scraper.Close()

	json := fmt.Sprintf("{\"url\": \"%s\"}", url)
	req := strings.NewReader(json)
	resp, err := http.Post("http://" + svc.scraper.GetListen(), "", req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != expectedServiceStatus {
		t.Fatalf(`Failed`)
	}

	serviceMetrics, err := svc.metrics.GetService(expectedServiceStatus)
	if err != nil || serviceMetrics != 1 {
		t.Fatalf(`Failed`)
	}

	scrapeMetrics, err := svc.metrics.GetScrapes(url, expectedScrapeStatus)
	if err != nil || scrapeMetrics != 1 {
		t.Fatalf(`Failed`)
	}
}

func TestWorkerWaitTimeout(t *testing.T) {
	args := Arguments{
		Listen: ":0",
		Workers: 0, // no workers, all requests will timeout =)
		Timeout: 1000,
		MetricsListen: ":0",
	}

	url := "http://it.doesnt.matter"
	expectedServiceStatus := http.StatusRequestTimeout

	svc := StartService(args)
	defer svc.metrics.Close()
	defer svc.scraper.Close()

	json := fmt.Sprintf("{\"url\": \"%s\"}", url)
	req := strings.NewReader(json)
	resp, err := http.Post("http://" + svc.scraper.GetListen(), "", req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != expectedServiceStatus {
		t.Fatalf(`Failed`)
	}

	serviceMetrics, err := svc.metrics.GetService(expectedServiceStatus)
	if err != nil || serviceMetrics != 1 {
		t.Fatalf(`Failed`)
	}

	workerWaitCount, err := svc.metrics.GetCountWorkerWaits()
	if err != nil || workerWaitCount != 1 {
		t.Fatalf(`Failed`)
	}

	workerWaitSum, err := svc.metrics.GetSumWorkerWaits()
	if err != nil || workerWaitSum < 1000 || workerWaitSum > 1010 {
		t.Fatalf(`Failed`)
	}
}

func TestMain(m *testing.M) {
	code := m.Run() 
	os.Exit(code)
}
