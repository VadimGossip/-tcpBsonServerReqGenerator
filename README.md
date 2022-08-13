BSON request generator
Позволяет генерировать N запросов в секунду на протяжении M секунд. Оба параметра задаются в конфиге.

Пример:
```
requestGenerator:
    routePerSec: 100
    workTimeSec: 5
```
Пример запроса:
```
type RouteRequest struct {
  RequestType string           `bson:"requestType"`
  MsgId	    int                `bson:"msgId"`
  SendTime    time.Time     	 `bson:"sendTime"`
  RouteDuration time.Duration	 `bson:"routeDur"`
}
```
Параметр RouteDuartion нужен для эмуляции как-то работы на стороне сервера. 

Пример сервера:
https://github.com/VadimGossip/tcpBsonServerExample

Пример ответа от сервера:
```
type RouteResponse struct {
  Err            string        `bson:"err,omitempty"`
  SendTime       time.Time     `bson:"sendTime,omitempty"`
  RouteStartTime time.Time		 `bson:"routeBegin,omitempty"`
  RouteEndTime   time.Time		 `bson:"routeEnd,omitempty"`
}
```
По завершении работы программа формирует ответ вида:
```
config {RoutePerSec:100 WorkTimeSec:5}
Route Request Sent 500
Route Histogramm map[200:500]
Route Request Avg Answer Duration 109.552003ms
Route Request Min Answer Duration 106.7175ms
Route Request Max Answer Duration 127.097ms
```
Шаг гистограммы 100ms
