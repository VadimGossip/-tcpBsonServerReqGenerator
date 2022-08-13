BSON request generator
Allows to generate N requests per second for M seconds. Both parameters are set in the config.

Example:
```
requestGenerator:
    routePerSec: 100
    workTimeSec: 5
```
Query example:
```
type RouteRequest struct {
  RequestType string           `bson:"requestType"`
  MsgId	    int                `bson:"msgId"`
  SendTime    time.Time     	 `bson:"sendTime"`
  RouteDuration time.Duration	 `bson:"routeDur"`
}
```
The RouteDuartion parameter is needed to emulate somehow working on the server side. 

Server example:
https://github.com/VadimGossip/tcpBsonServerExample

Server response example:
```
type RouteResponse struct {
  Err            string        `bson:"err,omitempty"`
  SendTime       time.Time     `bson:"sendTime,omitempty"`
  RouteStartTime time.Time		 `bson:"routeBegin,omitempty"`
  RouteEndTime   time.Time		 `bson:"routeEnd,omitempty"`
}
```
When the program finishes, it generates a response of the form:
```
config {RoutePerSec:100 WorkTimeSec:5}
Route Request Sent 500
Route Histogramm map[200:500]
Route Request Avg Answer Duration 109.552003ms
Route Request Min Answer Duration 106.7175ms
Route Request Max Answer Duration 127.097ms
```
Histogram Step 100ms
