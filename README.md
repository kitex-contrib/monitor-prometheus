# Prometheus monitoring for Kitex
English | [中文](README_ZH.md)

## Abstract
The approximate workflow of Prometheus:
1. The Prometheus server periodically pulls metrics from configured jobs or exporters (pull mode), or receives metrics pushed from Pushgateway (push mode), or fetches metrics from other Prometheus servers.
2. The Prometheus server locally stores the collected metrics, runs defined `alert.rules`, records new time series, or sends alerts to Alertmanager.
3. Alertmanager processes received alerts according to the configuration file and issues alarms.
4. In the graphical interface, visualizes collected data, for example, integrating with Grafana.

### Data Model
The data stored in Prometheus consists of time series, uniquely identified by a metric's name and a series of labels (key-value pairs), where different labels represent different time series.
- **name:** Typically represents the functionality of the metric; note that metric names consist of ASCII characters, digits, underscores, and colons and must adhere to the regular expression `[a-zA-Z_:][a-zA-Z0-9_:]*`.
- **tag:** Identifies feature dimensions for filtering and aggregation. For example, PSM and method information. Tag keys consist of ASCII characters, digits, and underscores and must adhere to the regular expression `[a-zA-Z_:][a-zA-Z0-9_:]*`.
- **sample:** Actual time series comprising a float64 value and a timestamp in milliseconds.
- **metric:** Represented in the following format: `<metric name>{<label name>=<label value>, ...}`

### Metric Types

#### Counter
- Understandable as an increment-only counter, typical applications include counting requests, completed tasks, occurring errors, etc.
- Corresponds to gopkg/metrics' EmitStore.

#### Gauge
- A standard metric; typical applications include counting goroutines.
- Can be increased or decreased arbitrarily.
- Corresponds to gopkg/metrics' EmitCounter.

#### Histogram
- Generates histogram data used for statistical analysis of sample distributions; typical applications include pct99, average CPU usage, etc.
- Allows sampling, grouping, and statistics on observed results.
- Corresponds to gopkg/metrics' EmitTimer.

#### Summary
- Similar to Histogram, providing count and sum functions for observed values.
- Offers percentile functionality, dividing tracked results by percentage.
- Summary's percentiles are calculated directly on the client-side, resulting in better performance when querying via PromQL. In contrast, Histogram consumes more resources; for clients, Histogram consumes fewer resources.

## Labels
- **type -** Request type
  - pingpong - Single request, single response
  - oneway - Single request, no response
  - streaming - Multiple requests, multiple responses
- **caller -** Requesting service name
- **callee -** Requested service name
- **method -** Request method name
- **status -** Status after a complete RPC:
  - succeed - Request successful
  - error - Request failed

## Metrics
- **Total number of requests handled by the Client:**
  - **Name:** kitex_client_throughput
  - **Tags:** type, caller, callee, method, status
- **Latency of request handling at the Client (Response received time - Request initiation time, in microseconds):**
  - **Name:** kitex_client_latency_us
  - **Tags:** type, caller, callee, method, status
- **Total number of requests handled by the Server:**
  - **Name:** kitex_server_throughput
  - **Tags:** type, caller, callee, method, status
- **Latency of request handling at the Server (Processing completion time - Request received time, in microseconds):**
  - **Name:** kitex_server_latency_us
  - **Tags:** type, caller, callee, method, status

## Useful Examples
For Prometheus query syntax, refer to [Querying basics | Prometheus](https://prometheus.io/docs/prometheus/latest/querying/basics/). Here are some commonly used examples:

**Client throughput of succeed requests**
```
sum(rate(kitex_client_throughput{status="succeed"}[1m])) by (callee,method)
```

**Client latency pct99 of succeed requests**
```
histogram_quantile(0.99,
sum(rate(kitex_client_latency_us_bucket{status="succeed"}[1m])) by (caller,callee,method,le)
)
```

**Server throughput of succeed requests**
```
sum(rate(kitex_server_throughput{status="succeed"}[1m])) by (code,callee,method)
```

**Server latency pct99 of succeed requests**
```
histogram_quantile(0.99,
sum(rate(kitex_server_latency_us_bucket{status="succeed"}[5m])) by (caller,callee,method,le)
)
```

**Pingpong request error rate**
```
sum(rate(kitex_server_throughput{status="error"}[1m])) by (status,callee,method)
```

## Usage Example

### Client

```go
import (
    "github.com/kitex-contrib/monitor-prometheus"
    kClient "github.com/cloudwego/kitex/client"
)

...
	client, _ := testClient.NewClient(
	"DestServiceName", 
	kClient.WithTracer(prometheus.NewClientTracer(":9091", "/kitexclient")))
	
	resp, _ := client.Send(ctx, req)
...
```

### Server

```go
import (
    "github.com/kitex-contrib/monitor-prometheus"
    kServer "github.com/cloudwego/kitex/server"
)

func main() {
...
	svr := api.NewServer(
	    &myServiceImpl{}, 
	    kServer.WithTracer(prometheus.NewServerTracer(":9092", "/kitexserver")))
	svr.Run()
...
}
```

## Visualization Interface
### Installing Prometheus
1. Refer to [Official Documentation](https://prometheus.io/docs/introduction/first_steps/), download and install the Prometheus server.
2. Edit prometheus.yml, modify the scrape_configs item:
```yaml
# Load rules once and periodically evaluate them according to the global 'evaluation_interval'.
rule_files:
# - "first_rules.yml"
# - "second_rules.yml"

# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
- job_name: 'kitexclient'
  scrape_interval: 1s
  metrics_path: /kitexclient
  static_configs:
  - targets: ['localhost:9091'] # scrape data endpoint
- job_name: 'kitexserver'
  scrape_interval: 1s
  metrics_path: /kitexserver
  static_configs:
  - targets: ['localhost:9092'] # scrape data endpoint
```
3. Start Prometheus:
```console
prometheus --config.file=prometheus.yml --web.listen-address="0.0.0.0:9090"
```
4. Access `http://localhost:9090/targets` in your browser to view the configured scrape nodes.

### Installing Grafana
1. Refer to the [official website](https://grafana.com/grafana/download), download and install Grafana.
2. Access `http://localhost:3000` in your browser; the default username and password are both `admin`.
3. Configure the data source by navigating to `Configuration` -> `Data Source` -> `Add data source`. After configuring, click on `Save & Test` to verify if it's functioning properly.
4. Create monitoring dashboards by going to `Create` -> `Dashboard`. Add metrics like throughput and pct99 based on your requirements. You can refer to the sample configurations provided in the "Useful Examples" section above.