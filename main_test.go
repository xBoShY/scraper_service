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

	url := "https://www.phaidra.ai/trackrecord"
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
		t.Fatalf(`Expected Status Code %d, got %d`, expectedServiceStatus, resp.StatusCode)
	}

	serviceMetrics, err := svc.metrics.GetService(expectedServiceStatus)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if serviceMetrics != 1 {
		t.Fatalf(`Expected service metrics to be 1, got %f`, serviceMetrics)
	}

	scrapeMetrics, err := svc.metrics.GetScrapes(url, expectedScrapeStatus)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if scrapeMetrics != 1 {
		t.Fatalf(`Expected scrape metrics to be 1, got %f`, scrapeMetrics)
	}
}

func TestServiceStatusRedirect(t *testing.T) {
	args := Arguments{
		Listen: ":0",
		Workers: 1,
		Timeout: 1000,
		MetricsListen: ":0",
	}

	url := "https://google.com"
	expectedServiceStatus := http.StatusOK
	expectedScrapeStatus := http.StatusMovedPermanently

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
		t.Fatalf(`Expected Status Code %d, got %d`, expectedServiceStatus, resp.StatusCode)
	}

	serviceMetrics, err := svc.metrics.GetService(expectedServiceStatus)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if serviceMetrics != 1 {
		t.Fatalf(`Expected service metrics to be 1, got %f`, serviceMetrics)
	}

	scrapeMetrics, err := svc.metrics.GetScrapes(url, expectedScrapeStatus)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if scrapeMetrics != 1 {
		t.Fatalf(`Expected scrape metrics to be 1, got %f`, scrapeMetrics)
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
		t.Fatalf(`Expected Status Code %d, got %d`, expectedServiceStatus, resp.StatusCode)
	}

	serviceMetrics, err := svc.metrics.GetService(expectedServiceStatus)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if serviceMetrics != 1 {
		t.Fatalf(`Expected service metrics to be 1, got %f`, serviceMetrics)
	}

	scrapeMetrics, err := svc.metrics.GetScrapes(url, expectedScrapeStatus)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if scrapeMetrics != 1 {
		t.Fatalf(`Expected scrape metrics to be 1, got %f`, scrapeMetrics)
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
	if err != nil {
		t.Fatalf(err.Error())
	}
	if serviceMetrics != 1 {
		t.Fatalf(`Expected Status Code %d, got %d`, expectedServiceStatus, resp.StatusCode)
	}

	workerWaitCount, err := svc.metrics.GetCountWorkerWaits()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if workerWaitCount != 1 {
		t.Fatalf(`Expected worker wait metrics to be 1, got %d`, workerWaitCount)
	}

	workerWaitSum, err := svc.metrics.GetSumWorkerWaits()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if workerWaitSum < 1000 || workerWaitSum > 1010 {
		t.Fatalf(`Expected worker wait timeout metrics to be between 1000 and 1010, got %f`, workerWaitSum)
	}
}

func TestMain(m *testing.M) {
	code := m.Run() 
	os.Exit(code)
}
