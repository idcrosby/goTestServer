goTestServer
============

Simple server in Go to simulate various backend responses.

Current endpoints:

Delay - wait a specified length of time before responding.
    Path: /delay 
    Parameters: sleep(milliseconds)
         
Return Status - return the specified HTTP Status code. 
  Path: /returnStatus 
  Parameters: status(number)

Sample data - respond with dummy data of either a) the specified size or b) for the specified duration. 
  Path: /sampleResponse 
  Parameters: time(milliseconds) latency(milliseconds) size(number of bytes)
         
Dump Rquest - respond with metadata from the request. 
  Path: /dumpRequest 
  Parameters: none
         
Add Response Header(s) - add the specified headers to the response. 
  Path: /addHeader 
  Parameters: name(String) value(String)
         
Cache test - serves generic content with some specific caching headers. 
  Path: /cacheTests/ 
  Parameterss: none
         
Serve Content - serves content from any file on the server. 
  Path: /getContent/ 
  Parameters: none
  
Validate JSON - validates and formats JSON data passed in the request body. 
  Path: /validateJson 
  Parameters: none
