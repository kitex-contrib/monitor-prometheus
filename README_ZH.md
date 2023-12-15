# Prometheus monitoring for Kitex
[English](README.md) | 中文

## Abstract
Prometheus 大概的工作流程：
1. Prometheus server 定期从配置好的 jobs 或者 exporters 中拉（pull模式） metrics，或者接收来自 Pushgateway 发过来（push模式）的 metrics，或者从其他的 Prometheus server 中拉 metrics；
2. Prometheus server 在本地存储收集到的 metrics，并运行已定义好的 `alert.rules`，记录新的时间序列或者向 Alertmanager 推送警报；
3. Alertmanager 根据配置文件，对接收到的警报进行处理，发出告警；
4. 在图形界面中，可视化采集数据，例如对接Grafana。

### 数据模型
Prometheus 中存储的数据为时间序列，是由 metric 的名字和一系列的标签（键值对）唯一标识的，不同的标签则代表不同的时间序列。
- name：一般用于表示 metric 的功能；注意，metric 名字由 ASCII 字符，数字，下划线，以及冒号组成，必须满足正则表达式 [a-zA-Z_:][a-zA-Z0-9_:]*；
- tag：标识了特征维度，便于过滤和聚合。例如 PSM 和 method 等信息。tag 中的 key 由 ASCII 字符，数字，以及下划线组成，必须满足正则表达式 [a-zA-Z_:][a-zA-Z0-9_:]*；
- sample：实际的时间序列，每个序列包括一个 float64 的值和一个毫秒级的时间戳；
- metric：通过如下格式表示：<metric name>{<label name>=<label value>, ...}

### Metric类型

#### Counter
- 可以理解为只增不减的计数器，典型的应用如：请求的个数，结束的任务数， 出现的错误数等等；
- 对应 gopkg/metrics 的 EmitStore。

#### Gauge
- 一种常规的 metric，典型的应用如：goroutines 的数量；
- 可以任意加减；
- 对应 gopkg/metrics 的 EmitCounter。

#### Histogram
- 生成直方图数据，用于统计和分析样本的分布情况，典型的应用如：pct99，CPU 的平均使用率等；
- 可以对观察结果采样，分组及统计。
- 对应 gopkg/metrics 的 EmitTimer。

#### Summary
- 类似于 Histogram，提供观测值的 count 和 sum 功能；
- 提供百分位的功能，即可以按百分比划分跟踪结果；
- Summary 的分位数是直接在客户端计算完成，因此对于分位数的计算而言，Summary 在通过 PromQL 进行查询时有更好的性能表现，而 Histogram 则会消耗更多的资源，对于客户端而言 Histogram 消耗的资源更少。

## Labels
- type - 请求类型
    - pingpong - 单次请求，单次应答
    - oneway - 单次请求，没有应答
    - streaming - 多次请求，多次应答
- caller - 请求方 service name
- callee - 被请求方 service name
- method - 请求的 method name
- status - 一次完整的 rpc 之后，返回的状态
    - succeed - 请求成功
    - error - 请求失败

## Metrics
- Client 端处理的请求总数：
    - Name: kitex_client_throughput
    - Tags: type, caller, callee, method, status
- Client 端请求处理耗时（收到应答时间 - 发起请求时间，单位 us）：
    - Name: kitex_client_latency_us
    - Tags: type, caller, callee, method, status
- Server 端处理的请求总数：
  - Name: kitex_server_throughput
  - Tags: type, caller, callee, method, status
- Server 端请求处理耗时（处理完请求时间 - 收到请求时间，单位 us）：
    - Name: kitex_server_latency_us
    - Tags: type, caller, callee, method, status

## Useful Examples
Prometheus 的查询语法可以参考 [Querying basics | Prometheus](https://prometheus.io/docs/prometheus/latest/querying/basics/), 这里给出一些常用示例：

**client throughput of succeed requests**
```
sum(rate(kitex_client_throughput{status="succeed"}[1m])) by (callee,method)
```

**client latency pct99 of succeed requests**
```
histogram_quantile(0.99,
sum(rate(kitex_client_latency_us_bucket{status="succeed"}[1m])) by (caller,callee,method,le)
)
```

**server throughput of succeed requests**
```
sum(rate(kitex_server_throughput{status="succeed"}[1m])) by (code,callee,method)
```

**server latency pct99 of succeed requests**
```
histogram_quantile(0.99,
sum(rate(kitex_server_latency_us_bucket{status="succeed"}[5m])) by (caller,callee,method,le)
)
```

**pingpong request error rate**
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

## 可视化界面
### 安装 Prometheus
1. 参考[官网](https://prometheus.io/docs/introduction/first_steps/), 下载并安装 Prometheus server
2. 编辑 prometheus.yml，修改 scrape_configs 项
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
3. 启动 Prometheus
```console
prometheus --config.file=prometheus.yml --web.listen-address="0.0.0.0:9090"
```

4. 浏览器访问 `http://localhost:9090/targets`, 可以看到刚才配置的抓取节点

### 安装 Grafana
1. 参考[官网](https://grafana.com/grafana/download) ，下载并安装 Grafana
2. 浏览器访问 `http://localhost:3000`, 账号密码默认都是 `admin`
3. 配置数据源 `Configuration` ->`Data Source` -> `Add data source`，配置后点击 `Save & Test` 测试验证是否生效
4. 添加监控界面 `Create` -> `dashboard`，根据自己的需求添加 throughput 和 pct99 等监控指标，可以参考上面 `Useful Examples` 给出的样例。