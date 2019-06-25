# What is Lagrande?

Lagrande is a load-testing tool for your TSDB. It is written in Go and it was designed to support many data formats, sending over UDP/TCP or HTTP and being highly configurable in regards to the metrics it's generating.

## Supported TSDB / formats:
* [Atlas](https://github.com/Netflix/atlas)
* [Carbon plaintext protocol](https://graphite.readthedocs.io/en/latest/feeding-carbon.html#the-plaintext-protocol)
* [InfluxDB](https://www.influxdata.com/products/influxdb-overview/)
* [M3DB](https://m3db.io/)

## Kinds of metrics:
* Integer fixed value
* Float fixed value
* Integer counter (increment/decrement)
* Float counter (increment/decrement)
* Integer random value
* Float random value
* Latency (random float generated from a Beta probability distribution)

# Getting started

## Quick-start

First start Graphite and Grafana containers so that you have a destination to send metrics to and a nice dashboard to plot your metrics:
```bash
docker run --rm -d --name graphite -p 80:80 -p 2003-2004:2003-2004 -p 8125:8125/udp -p 8126:8126 graphiteapp/graphite-statsd
docker run --rm -d --name grafana -p 3000:3000 grafana/grafana
```

Then grab one of the pre-built [releases](#pre-built-binary-releases), move it to one of the folders in your `$PATH` and just launch it: `lagrande`.

## Installation

### Pre-built binary releases

Pre-built binairies will be made available as [GitHub releases](https://github.com/aleveille/lagrande/releases).

### Building from source

Make sure your local Golang installation is [properly set up](https://golang.org/doc/install) and then run:
```bash
go get github.com/aleveille/lagrande
go install github.com/aleveille/lagrande
```

## Configuration

The command line arguments listed below are also documented with the usual Go `-h` flag. Only the most important flags will be documented here:

|Flag|Default value|Accepted values|Description|
|-|-|-|-|
|`-endpoint`|`<empty>`|`<URI string>`|The endpoint to send the data to, must be an URI.|
|`-format`|`carbon`|`carbon`, `influxdb`, `atlas`|The data format. Some TSDB support more than one format.|
|`-protocol`|`auto`|`auto`, `http`, `tcp`, `udp`|Auto will automatically pick an appropriate protocol based on the format (eg: HTTP for Atlas and TCP for Carbon) Not all formats support all protocols!|
|`-profile`|`'counterInt={name: fixedValue, value: 10, increment: 0},randomInt={name: jiggle, min: 50, max: 75}`|`<string>`|The configuration for the generator(s) to use. See [Metric generation reference](#metric-generation-reference).|
|`-interval`|`1s`|`<Go duration string>`|How often each worker will generate metrics.|
|`-metricNamespacePrefix`|`lagrande.`|`<string>`|For namespacing metrics, this will be prepended to the metric name. Support placeholders: NODENAME, WORKERNUM, WORKERFULLNAME|
|`-metricNamespaceSuffix`|`-WORKERNUM`|`<string>`|For namespacing metrics, this will be appended to the metric name. Support placeholders: NODENAME, WORKERNUM, WORKERFULLNAME|
|`-tags`|`node=NODENAME,process=lagrande,thread=WORKERFULLNAME`|`<string>`|Comma-delimited list of tags of format name=value. Supports placeholders: NODENAME, PID, WORKERNUM, WORKERFULLNAME, METRICNAME.|
|`-workersCount`|`10`|`<URI>`|Number of parallel workers that will send metrics.|
|`-workersInterval`|`1s`|`<Go duration string>`|Wait time between starting workers, must be a >= 0 Go Duration.|

### Metric generation reference

You can use `-profile` flag to provide inline configuration on which generators to create and how to configure them. 

The general syntax is `generatorType={[key: value][, ...]}[, ...]`. Spaces are optional, but help in reading the configuration.

Eg: `counterInt={name: metric1, value: 10}, randomInt={name: random, min: 10, max: 20}`

|Generator type|Config elements|
|-|-|
|counterInt|<ul><li>`name`: name of the metric</li><li>`increment`: increment or decrement the value each time a metric is generated. If 0, counterInt will be a fixed number</li><li>`max`: the maximum value of the counter</li><li>`min`: the minimum value of the counter</li><li>`reset`: whehter to reset the counter to min (or max) when it gets to max (or min), based on the sign of `increment`</li><li>`value`: the initial value of the counter</li></ul>|
|counterFloat|<ul><li>`name`: name of the metric</li><li>`increment`: increment or decrement the value each time a metric is generated. If 0, counterFloat will be a fixed number</li><li>`max`: the maximum value of the counter</li><li>`min`: the minimum value of the counter</li><li>`reset`: whehter to reset the counter to min (or max) when it gets to max (or min), based on the sign of `increment`</li><li>`value`: the initial value of the counter</li></ul>|
|latency|<ul><li>`name`: name of the metric</li><li>`alpha`: the alpha parameter of the Gamma distribution. You can think of alpha as the skewness. Data is gathered more around the left with lower values of Alpha and more to the right with higher values of Alpha.</li><li>`beta`: the beta parameter of the Gamma distribution. You can think of beta as the control to how much the data is grouped or scattered. Data is more grouped around the "peak" with lower values of Beta and more scattered (longer and bigger tail) with higher values of Beta.</li><li>`max`: the maximum value generated</li><li>`min`: the minimum value generated</li></ul>|
|randomInt|<ul><li>`name`: name of the metric</li><li>`max`: the maximum value of the random integer</li><li>`min`: the minimum value of the random integer</li></ul>|
|randomFloat|<ul><li>`name`: name of the metric</li><li>`max`: the maximum value of the random float</li><li>`min`: the minimum value of the random float</li></ul>|

#### Examples

##### Fixed (static) integer metric

This profile will create a metric with a value of 42 that never changes. It doesn't require any computation and is therefore very fast. A fixed value counterInt generator can be very good if you're interested in generating a very high number of metrics and you aren't much interested in their variance.
```
lagrande -profile 'counterInt={name: staticValue, value: 42, increment: 0}'
```

##### Integer counter that counts from 0 to 100000

Integer counters with a low cardinality (<500000) gets their string byte representation cached. This counter is also very efficient. When this counter gets to 100000, it will reset its value to 0.
```
lagrande -profile 'counterInt={name: counter, value: 0, increment: 1, maximum: 100000}'
```

##### Integer counter that counts from 0 to math.MaxInt32 and plateau at that value

By using the `reset: false` configuration parameter for the generator, you can have the counter plateau at its final value.
```
lagrande -profile 'counterInt={name: biggerCounter, value: 0, increment: 1, reset: false}'
```

##### Float counter that decrement 

The increment can be a negative number. This counter will decrease from 1000.00 to 0.00 and will stay at 0.00 when it gets there.
```
lagrande -profile 'counterFloat={name: floatCountdown, value: 1000, increment: -2.5, minimum: 0, reset: false}'
```

##### Random int between 10 and 100

Random int can be good to simulate discrete values such as the number of users on a website. This is a gauge-type metric.
```
lagrande -profile 'randomInt={name: connectedUsers, min: 10, max: 200}'
```

##### Random float between 0.00 and 1.00

Random float can be good to simulate continuous values such as percentages. This is a gauge-type metric.
```
lagrande -profile 'randomFloat={name: someBufferUsage, min: 0, max: 1}'
```

##### Latency like value

The `latency` generator generates a random number between its `min` and `max` configuration parameters, but the generation is done with a gamma probability distribution. This means most values will be close to the lower bound (min parameter) and higher values have a decreased chance of being generated.
```
lagrande -profile 'latency={name: requestTime, min: 150, max: 8000, alpha: 1.5, beta: 10}'
```

##### Multiple generators

It is possible more than one generator. The number of metrics per seconds will be: (Number of workers * Number of generator) / Interval.
```
lagrande -profile 'counterInt={name: staticValue, value: 42, increment: 0}, counterInt={name: counter, value: 0, increment: 1, maximum: 100000}, randomInt={name: connectedUsers, min: 10, max: 200}, randomFloat={name: someBufferUsage, min: 0, max: 1}'
```

## List of convenient Docker containers

**TSDB:**
* Graphite: `docker run --rm -d --name graphite -p 80:80 -p 2003-2004:2003-2004 -p 8125:8125/udp -p 8126:8126 graphiteapp/graphite-statsd`
* InfluxDB: `docker run --rm -d --name influxdb -p 8086:8086 influxdb`
* IRONdb: `docker run --rm -d --name irondb -p 2003-2004:2003-2004 -p 8112:8112 irondb/irondb`
* M3DB: `docker run --rm -d --name m3db -p 7201:7201 -p 7203:7203 -p 9003:9003 -v $HOME/tmp/m3db_data:/var/lib/m3db --privileged quay.io/m3db/m3dbnode:latest`

**Support tooling:**
* Grafana: `docker run --rm -d --name grafana -p 3000:3000 grafana/grafana`

# Contributing

If you have a feature request or a question, feel free to open an issue. Otherwise, I accept contributions via GitHub pull requests.

# License

MIT License

Copyright (c) 2019 Alexandre Léveillé

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.