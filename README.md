fluxrus
===
[logrus](https://github.com/Sirupsen/logrus) hook for [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/)

usage
---
###Installation
```
go get github.com/HeavyHorst/fluxrus
```

###Example
```go
import (
    "github.com/HeavyHorst/fluxrus"
	"github.com/Sirupsen/logrus"
)

func main() {
    log := logrus.New()
	hook, err := fluxrus.New("http://influxdbhost:8086", "mydb", "my_measurement", fluxrus.WithBatchSize(2000), fluxrus.WithTags([]string{"mytag"}))
	if err == nil {
		log.Hooks.Add(hook)
	}
	defer hook.Flush()
}
```

Details
---
###Buffer
All log messages are buffered by default to increase the performance. You need to run hook.Flush() on program exit to get sure that all messages are written to the database.

###Configuration
The constructor takes some optional configuration options:

 1. WithBatchInterval(int):  flush the buffer at the given interval in seconds. (default is 5s)
 2. WithBatchSize(int): Max size of the buffer. The buffer gets flushed to the db if this threshold is reached. (default is 200)
 3. WithPrecision(string): Sets the timestamp precision. (default is ns)
 4. WithTags([]string): Use some logrus fields as influxdb tags instead of influxdb fields.
 5. WithClient(influx.Client): A custom influxdb client.

### Message Field
We will insert your message into InfluxDB with the field message.
