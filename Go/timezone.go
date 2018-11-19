package main

import (
	"fmt"
	"time"
	"os"
	"encoding/json"
	"log"
	"net/http"
	"html/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strings"
)

var apikey = os.Getenv("APIKEY")

// Template files
var in, err = template.ParseFiles("index.html")
var ti, tierr = template.ParseFiles("time.html")

// Metrics (Prometheus)
var (
	Hits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_hits_total",
			Help: "Number of hits",
		})
)

var (
	GeoLatency = promauto.NewSummary(
		prometheus.SummaryOpts{
			Name:"api_geo_latency_seconds",
			Help:"TODO",
		})
)

var (
	TimezoneLatency = promauto.NewSummary(
		prometheus.SummaryOpts{
			Name:"api_timezone_latency_seconds",
			Help:"TODO",
		})
)


func index(w http.ResponseWriter, r *http.Request) {
	// Number of requests 
	Hits.Inc()
	// Return index if method is not POST (form)
	if r.Method != http.MethodPost {
			in.Execute(w, nil)
			return
	}

	// Get city from form
	location := r.FormValue("city")
	if location == "" {
		fmt.Fprintf(w, "Please input city")
	}

	location = strings.Replace(location, " ", "+", -1)


	// Create API URL 
	var url = "https://maps.googleapis.com/maps/api/geocode/json?address=" + location + "&language=en&key=" + apikey
	
	
	// Query API (calculating latency)
	geostart := time.Now()
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	geofinal := time.Since(geostart)
	GeoLatency.Observe(geofinal.Seconds())

	// Struct to match JSON data
	type Response struct {
		Results []struct {
				Geometry struct {
						Location struct {
							Lat float64
							Lng float64
						}
				}
				Formatted_address string
		}
		Status string	
	}
	
	var data Response

	// Decode JSON 
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println(err)
	}		
	resp.Body.Close()

	// Check API response
	if data.Status != "OK" {
		fmt.Fprintf(w, "Unable to find city (API)")
		return
	}
	
	// Store data 
	lat := data.Results[0].Geometry.Location.Lat
	lng := data.Results[0].Geometry.Location.Lng
	name := data.Results[0].Formatted_address

	// Get current time in unix epoch format (UTC)
	var unixutcnow int = int(time.Now().UTC().Unix())
	
	// Convert to string and add to URL
	slat := fmt.Sprintf("%f", lat)
	slng := fmt.Sprintf("%f", lng)
	sunix := fmt.Sprintf("%d", unixutcnow)
	timeurl := "https://maps.googleapis.com/maps/api/timezone/json?location=" + slat + "," + slng + "&timestamp=" + sunix + "&key=" + apikey
	
	// Query API (calculating latency)
	timestart := time.Now()
	
	resp, err = http.Get(timeurl)
	if err != nil {
		fmt.Println(err)
	}
	
	timefinal := time.Since(timestart)
	TimezoneLatency.Observe(timefinal.Seconds())

	// Struct to match JSON data
	type Timeresponse struct {
		DstOffset int
		RawOffset int
		Status string
	}

	var tdata Timeresponse

	// Query API
	if err := json.NewDecoder(resp.Body).Decode(&tdata); err != nil {
		fmt.Println(err)
	}		
	resp.Body.Close()

	// Check API response
	if tdata.Status != "OK" {
		fmt.Fprintf(w, "Unable to get timezone (API)")
		return
	}
	// Store data
	dst := tdata.DstOffset
	raw := tdata.RawOffset

	// Add offset to obtain local time (UTC)
	timeunix := int64((unixutcnow + ((raw) + (dst))))
	t := time.Unix(timeunix, 0)
	z := t.UTC()
	
	// Struct to store results
	type datainfo struct {
		Name   string
		Hour string
		Min string
		Day int
		Month string
		Year int
	}
	// Store (and format) results
	infotime := datainfo {
		Name:   name,
		Hour: z.Format("15"),
		Min: z.Format("04"),
		Day: z.Day(),
		Month: z.Format("January"),
		Year: z.Year(),
	}	
	// Returns results to client
	ti.Execute(w, infotime)	
	
}	

func main() {
	// Check if API key is set
	if apikey == "" {
	fmt.Println("API Key error")
	return 
	}
	// Define routes for app and Prometheus (/metrics), starting web server on port 8080
 	
 	http.HandleFunc("/", prometheus.InstrumentHandlerFunc("/", index))	
 	http.Handle("/metrics", promhttp.Handler())
 	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
    log.Fatal(http.ListenAndServe(":8080", nil))
    
}    