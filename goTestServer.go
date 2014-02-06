package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"fmt"
	"time"
	"bytes"
	"strconv"
	"text/template"
	"log"
	"os"
	"encoding/json"
	"flag"
	"strings"
)

// Global variables
var updatedTime time.Time
var modTime time.Time = time.Now()
var InfoLog *log.Logger
var ErrorLog *log.Logger
var Verbose bool

func main() {
	// Define flags
	flag.BoolVar(&Verbose, "verbose", false, "Turn on verbose logging.")
	flag.Parse()

	// init loggers
	logFile, _ := os.Create("goServerLog.log")
	errorFile, _ := os.Create("goErrorLog.log")
	InfoLog = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	ErrorLog = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime)

	http.HandleFunc("/", errorHandler(defaultHandler))
	http.HandleFunc("/delay", errorHandler(delayHandler))
	http.HandleFunc("/returnStatus", errorHandler(returnStatusHandler))
	http.HandleFunc("/sampleResponse", errorHandler(sampleResponseHandler))
	http.HandleFunc("/addHeader", errorHandler(addHeaderHandler))
	http.HandleFunc("/dumpRequest", errorHandler(dumpRequestHandler))
	http.HandleFunc("/cacheTests/", errorHandler(cacheTestHandler))
	http.HandleFunc("/getContent/", errorHandler(contentHandler))
	http.HandleFunc("/validateJson", errorHandler(validateJsonHandler))

	http.ListenAndServe(":8089", nil)
}

func defaultHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("defaultHandler called")
	var mainTemplate, err = template.ParseFiles("main.html")
	check(err)
	mainTemplate.Execute(rw, nil)
}

func delayHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("delayHandler called")
	sleepString := retrieveParam(req, "sleep")
	sleepTime, err := time.ParseDuration(sleepString + "ms")
	check(err)
	delay(sleepTime)
}

func returnStatusHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("returnStatusHandler called")
	statusCode, err := strconv.Atoi(retrieveParam(req, "status"))
	check(err)
	setResponseStatus(statusCode, rw)
}

func sampleResponseHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("sampleResponseHandler called")
	duration, error := time.ParseDuration(retrieveParam(req, "time") + "ms")
	latency, err := time.ParseDuration(retrieveParam(req, "latency") + "ms")
	check(err)
	if error == nil {
		outputDotsByTime(duration, latency, rw)
	} else {
		size, err := strconv.Atoi(retrieveParam(req, "size"))
		check(err)
		outputDotsBySize(size, latency, rw)
	}
}

func addHeaderHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("addHeaderHandler called")
	names := strings.Split(retrieveParam(req, "name"), ",")
	values := strings.Split(retrieveParam(req, "value"), ",")
	for count, _ := range names {
		addHeader(rw, names[count], values[count])
	}
}

func dumpRequestHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("dumpRequestHandler called")
	rw.Write(requestAsString(req))
}

func cacheTestHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("cacheTestHandler called")
	addHeader(rw, "Cache-Control", "max-age=10")
	contentHandler(rw, req)
}

func contentHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("contentHandler called")
	fileName := req.URL.Path[12:]
	if fileName == "" {
		fileName = "sampleData.json"
	}
	fi, err := os.Open(fileName)
	if err != nil { 
		rw.WriteHeader(404)
		return
	}
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	http.ServeContent(rw, req, "sampleData", modTime, fi)
}

func validateJsonHandler(rw http.ResponseWriter, req *http.Request) {
	InfoLog.Println("validateJsonHandler called.")
	var buffer bytes.Buffer
	buffer.ReadFrom(req.Body)
	jsonData := validateJson(buffer.Bytes())
	if jsonData == nil {
		rw.WriteHeader(400)
	} else {
		rw.Write(jsonData)
	}
}

// Error Handler Wrapper
func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if Verbose {
			InfoLog.Println(string(requestAsString(req)[:]))
		}
		defer func() {
			if e, ok := recover().(error); ok {
				w.WriteHeader(500)
				ErrorLog.Println(e)
			}
		}()
		fn(w, req)
	}
}

func check(err error) { if err != nil { panic(err) } }

// Helper Methods

// Retreive parameters passed in via query or post body
func retrieveParam(req *http.Request, param string) string {
	params, err := url.ParseQuery(req.URL.RawQuery)
	check(err)
	value := params[param]

	if len(value) < 1 {
		return ""
	} else {
		return value[0]
	}
	// TODO read from POST body
}

// Create a string which contains headers passed in the request
func getHeadersAsString(header http.Header) []byte {
	var buffer bytes.Buffer
	for key, value := range header {
		buffer.WriteString(key)
		buffer.WriteString(":")
		buffer.WriteString(value[0])	
		buffer.WriteString("\n")
	}
	return buffer.Bytes()
}

// Create a string which contains the cookies passed in the request
func getCookiesAsString(request *http.Request) []byte {
	var buffer bytes.Buffer
	for _, cookie := range request.Cookies() {
		buffer.WriteString(cookie.Name)
		buffer.WriteString(" : ")
		buffer.WriteString(cookie.Value)
		buffer.WriteString("\n")
	}
	return buffer.Bytes()
}

// Create a string which contains all important request data
func requestAsString(request *http.Request) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	buffer.WriteString("Current Time: ")
	buffer.WriteString(time.Now().String())
	buffer.WriteString("\n")
	requestBytes, err := httputil.DumpRequest(request, true)
	check(err)
	buffer.Write(requestBytes)

	return buffer.Bytes()
}

// Add a header to the response with the specified name and value
func addHeader(rw http.ResponseWriter, name, value string) {
	rw.Header().Set(name, value)
}

// Sleep for specified time
func delay(sleep time.Duration) {
	time.Sleep(sleep)
}

// Output dots for the specified length of time
//  optional latency between outputs
func outputDotsByTime(duration, latency time.Duration, rw http.ResponseWriter) {
	start := time.Now()
	expired := false
	for !expired {
		rw.Write([]byte("."))
		if _, ok := rw.(http.Flusher); ok {
			rw.(http.Flusher).Flush()
		}
		time.Sleep(latency)
		if time.Now().Sub(start) > duration {
			expired = true
		}
	}
}

// Output as many dots specified by the size param
//  optional latency between dots 
func outputDotsBySize(size int, latency time.Duration, rw http.ResponseWriter) {
	if _, ok := rw.(http.Flusher); ok {
		// fmt.Println("flusher is valid")
	}
	for i := 0; i < size; i++ {
		fmt.Fprintf(rw, ".")
		if _, ok := rw.(http.Flusher); ok {
			rw.(http.Flusher).Flush()
		}
		time.Sleep(latency)
	}
}

// Set a specific status code on the response
func setResponseStatus(statusCode int, rw http.ResponseWriter) {
	rw.WriteHeader(statusCode)
}

// Verify the JSON post body is valid
//  return formatted JSON if it is valid
// func validateJson(body io.Reader) (buf []byte) {
func validateJson(bytes []byte) (buf []byte) {
	var f interface{}
	err := json.Unmarshal(bytes, &f)
	if err != nil {
		fmt.Printf("Error reading Json")
		return nil
	}
	buf, err = json.MarshalIndent(&f, "", "   ")
	check(err)
	return
}
