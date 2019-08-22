package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aleveille/lagrande/formatter"
	"github.com/aleveille/lagrande/generator"
	"github.com/aleveille/lagrande/metric"
	"github.com/aleveille/lagrande/publisher"
)

// TODO:
//   Support prometheus format (pushing to pushgateway)
//   Support histogram/summary (Prometheus)
//   New generator: CPU-like
//   New generator: Memory-like
//   Control channel(s)
//   Potentially: Support destroying && creating workers over time (to simulate hosts going down, coming up)

var (
	version   string
	buildDate string

	// CLI flags
	hostname              string
	endpoint              string
	format                string
	protocol              string
	profile               string
	logLevel              string
	dryRun                bool
	interval              string
	nodeName              string
	metricNamespacePrefix string
	metricNamespaceSuffix string
	tags                  string
	versionFlag           bool
	workersCount          int
	workersInterval       string

	// Variables computed from CLI flags
	generatorsArr           []generator.Generator
	intervalDuration        time.Duration
	workersIntervalDuration time.Duration
	sharedTags              string
	workersTags             string

	localFormatter        formatter.Formatter
	stringPid             string
	statsPrintToPushRatio = int(math.Round(float64(statsPrintInterval.Seconds()) / float64(statsPushInterval.Seconds())))
)

const (
	statsPushInterval  = 500 * time.Millisecond
	statsPrintInterval = 30 * time.Second
)

type emissionStat struct {
	workerNum          int
	successfullySent   int64
	unsuccessfullySent int64
	duration           time.Duration
}

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil || len(hostname) == 0 {
		hostname = "local"
	}
	hostname = strings.ToLower(hostname)
	// TODO make sure that hostname is [[:word:]] compliant

	stringPid = strconv.Itoa(os.Getpid())

	flag.StringVar(&endpoint, "endpoint", "", "Endpoint to publish metrics to")
	flag.StringVar(&format, "format", "carbon", "Publish format: \"atlas\",\"carbon\", \"influxdb\", \"m3db\" or \"timescale\"")
	flag.StringVar(&protocol, "protocol", "auto", "Publish protocol: \"auto\", \"http\", \"tcp\" or \"udp\". NB: not all format support all protocol!")
	flag.StringVar(&profile, "profile", "counterInt={name: fixedValue, value: 10, increment: 0},randomInt={name: jiggle, min: 50, max: 75}", "")
	flag.StringVar(&logLevel, "logLevel", "info", "Log level: \"trace\", \"debug\", \"info\", \"warn\", \"error\", \"fatal\", \"panic\"")
	flag.BoolVar(&dryRun, "dry-run", false, "Don't send any metrics")
	flag.StringVar(&interval, "interval", "1s", "Generate metrics every X unit of time, must be a > 0 Go Duration")
	flag.StringVar(&metricNamespacePrefix, "metricNamespacePrefix", "lagrande.", "How to namespace metrics. Eg: 'lagrande.mymetric'. Support text and placeholders: NODENAME, PID, WORKERNUM, WORKERFULLNAME")
	flag.StringVar(&metricNamespaceSuffix, "metricNamespaceSuffix", "-WORKERNUM", "How to namespace metrics. Eg: 'mymetric-6'. Support text and placeholders: NODENAME, PID, WORKERNUM, WORKERFULLNAME")
	flag.StringVar(&tags, "tags", "", "Comma-delimited list of tags of format name=value. Support placeholders: NODENAME, PID, WORKERNUM, WORKERFULLNAME, METRICNAME") // If defaulting to 'node=NODENAME,process=lagrande,thread=WORKERFULLNAME', make sure it plays nice with TSDB that don't support tags
	flag.BoolVar(&versionFlag, "version", false, "Print version information")
	flag.IntVar(&workersCount, "workers", 10, "Number of parallel workers that will send metrics")
	flag.StringVar(&workersInterval, "workersInterval", "1s", "Wait time between starting workers, must be a >= 0 Go Duration")
}

func main() {
	flag.Parse()

	if versionFlag {
		fmt.Printf("Lagrande version %s\nBuild date: %s\n", version, buildDate)
		os.Exit(0)
	}

	switch logLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	}

	err := processCliConfiguration()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	printConfig()
	// At this point we're done with parsing & validating the CLI configuration. Congrats!

	// Stats channel
	statsChan := make(chan emissionStat, workersCount*(statsPrintToPushRatio+1))
	go handleStats(statsChan)

	workersSpawnTicker := time.Tick(workersIntervalDuration)
	workersSpawnCount := 0
	//controlChan := make(chan bool, workersCount*(StatsPrintToPushRatio+1))

	for {
		select {
		// Need a control channel here
		case <-workersSpawnTicker:
			if workersSpawnCount < workersCount {
				// TODO: give a control channel for the workers to stop working
				go spawnWorker(workersSpawnCount, statsChan)
				log.Infof("Launched worker-%s-%d", stringPid, workersSpawnCount)
				workersSpawnCount++
			} else {
				log.Info("All workers launched")
				workersSpawnTicker = nil
			}
		}
	}

	// TODO check controlChan, set workersSpawnTicker to nil if we stop
}

func processCliConfiguration() error {
	var err error

	err = processFormatAndProtocol()
	if err != nil {
		return err
	}

	if dryRun {
		protocol = "dry-run"
	}

	intervalDuration, err = time.ParseDuration(interval)
	if err != nil || intervalDuration.Nanoseconds() <= int64(0) {
		return errors.New("Invalid interval specified. Make sure it's a duration greater than 0 and parsable by Go library: https://golang.org/pkg/time/#ParseDuration")
	}

	workersIntervalDuration, err = time.ParseDuration(workersInterval)
	if err != nil || workersIntervalDuration.Nanoseconds() < int64(0) {
		return errors.New("Invalid workersInterval specified. Make sure it's a duration greater or equal than 0 and parsable by Go library: https://golang.org/pkg/time/#ParseDuration")
	}

	err = processTags()
	if err != nil {
		return err
	}

	// TODO validate tags string
	metricNamespacePrefix = strings.ReplaceAll(metricNamespacePrefix, "NODENAME", hostname)
	metricNamespaceSuffix = strings.ReplaceAll(metricNamespaceSuffix, "NODENAME", hostname)
	metricNamespacePrefix = strings.ReplaceAll(metricNamespacePrefix, "PID", stringPid)
	metricNamespaceSuffix = strings.ReplaceAll(metricNamespaceSuffix, "PID", stringPid)

	err = processGenerators()
	if err != nil {
		return err
	}

	return nil
}

func processFormatAndProtocol() error {
	if protocol != "auto" && protocol != "tcp" && protocol != "udp" && protocol != "http" && protocol != "null" && protocol != "log" {
		return errors.New("The specified protocol is invalid")
	}

	if protocol == "null" || protocol == "log" {
		// Todo: validate format anyway
		return nil
	}

	// Validate format --> Protocol combinations and initialize formatters accordingly
	switch format {
	case "atlas":
		if protocol != "http" && protocol != "auto" {
			return errors.New("Only the HTTP protocol is supported with Atlas")
		}

		localFormatter = formatter.NewAtlasFormatter()

		if len(endpoint) == 0 {
			endpoint = "http://127.0.0.1:7101/api/v1/publish"
		}
		if protocol == "auto" {
			protocol = "http"
		}
	case "carbon":
		if protocol == "http" {
			return errors.New("The HTTP protocol isn't supported with Carbon")
		}

		localFormatter = formatter.NewCarbonFormatter()

		if len(endpoint) == 0 {
			endpoint = "127.0.0.1:2003"
		}
		if protocol == "auto" {
			protocol = "tcp"
		}
	case "influxdb":
		if protocol != "http" && protocol != "auto" {
			return errors.New("Only the HTTP protocol is supported with InfluxDB")
		}

		localFormatter = formatter.NewInfluxdbFormatter()

		if len(endpoint) == 0 {
			endpoint = "http://127.0.0.1:8086/write?db=mydb"
		}
		if protocol == "auto" {
			protocol = "http"
		}
	case "m3db":
		if protocol != "http" && protocol != "auto" {
			return errors.New("Only the HTTP protocol is supported with M3DB")
		}

		localFormatter = formatter.NewM3DBFormatter()

		if len(endpoint) == 0 {
			endpoint = "http://localhost:9003/writetagged"
		}
		if protocol == "auto" {
			protocol = "http"
		}
	case "timescale":
		if protocol != "auto" {
			return errors.New("Only the \"auto\" protocol (PostgreSQL:Timescale) is supported with Timescale")
		}
		protocol = "timescale"
		localFormatter = formatter.NewTimescaleFormatter()

		if len(endpoint) == 0 {
			endpoint = "postgres://postgres:postgresqltimescaletests@localhost/tsdb?sslmode=disable"
		}
	default:
		return errors.New("The specified format is invalid")
	}

	return nil
}

func processTags() error {
	tagsRE := regexp.MustCompile(`^[[:word:]]+=[[:word:]]+(,[[:word:]]+=[[:word:]]+)*$`)
	tagTokenizerRE := regexp.MustCompile(`[[:word:]]+=[[:word:]]+`)

	matched := len(tags) == 0 || tagsRE.MatchString(tags)
	if matched == false {
		return errors.New("Error while validating tags string. Please make sure it's a comma-delimited strings of key=value without any space and 'word' character class for both key and value (eg: tag1=value1,tag2=WORKERNUM)")
	}

	// Supported placeholders for tags: NODENAME, WORKERNUM, WORKERFULLNAME, METRICNAME
	// WORKERNUM and WORKERFULLNAME will be saved into workersTags
	// METRICNAME will be saved into metricsTags
	// Anything else will be saved into sharedTags
	tags = strings.ReplaceAll(tags, "NODENAME", hostname)
	tags = strings.ReplaceAll(tags, "PID", stringPid)

	for _, m := range tagTokenizerRE.FindAllString(tags, -1) {
		if strings.Contains(m, "WORKERNUM") || strings.Contains(m, "WORKERFULLNAME") || strings.Contains(m, "METRICNAME") {
			appendTag(&workersTags, &m)
		} else {
			appendTag(&sharedTags, &m)
		}
		log.Tracef("found: %s\n", m)
	}

	return nil
}

func appendTag(tagList *string, tag *string) {
	if len(*tagList) > 0 {
		*tagList = fmt.Sprintf("%s,%s", *tagList, *tag)
	} else {
		*tagList = *tag
	}
}

func processGenerators() error {
	generatorsRE := regexp.MustCompile(`[a-zA-Z]*={[^}]*}(,[a-zA-Z]*={[^}]*})*`)
	generatorTokenizerRE := regexp.MustCompile(`[a-zA-Z]*={[^}]*}`)
	argumentsRE := regexp.MustCompile(`[a-zA-Z]*:\s?[^,}]*`)

	matched := generatorsRE.MatchString(profile)
	if matched == false {
		return errors.New("Error while validating profile string. Please refer to the doc and examples")
	}

	cliGenerators := generatorTokenizerRE.FindAllString(profile, -1)
	generatorsArr = make([]generator.Generator, len(cliGenerators), len(cliGenerators))

	for i, gen := range cliGenerators {
		var err error
		generatorName := strings.Split(gen, "=")[0]
		cliArguments := argumentsRE.FindAllString(gen, -1)

		switch generatorName {
		case "counterInt":
			generatorsArr[i], err = generator.NewIntCounterGenerator(generator.CLIConfig{Args: cliArguments}, localFormatter.FormatTags(&sharedTags), &localFormatter)
		case "counterFloat":
			generatorsArr[i], err = generator.NewFloatCounterGenerator(generator.CLIConfig{Args: cliArguments}, localFormatter.FormatTags(&sharedTags), &localFormatter)
		case "latency":
			generatorsArr[i], err = generator.NewLatencyDistributionGenerator(generator.CLIConfig{Args: cliArguments}, localFormatter.FormatTags(&sharedTags), &localFormatter)
		case "randomInt":
			generatorsArr[i], err = generator.NewIntRandomGenerator(generator.CLIConfig{Args: cliArguments}, localFormatter.FormatTags(&sharedTags), &localFormatter)
		case "randomFloat":
			generatorsArr[i], err = generator.NewFloatRandomGenerator(generator.CLIConfig{Args: cliArguments}, localFormatter.FormatTags(&sharedTags), &localFormatter)
		default:
			return errors.New("Invalid generatorName in the profile string, please refer to the doc")
		}

		if err != nil {
			log.Errorf("Error while instanciating %s generator:\n%s", generatorName, err)
		}
	}

	return nil
}

func printConfig() {
	// Configuration output
	log.Infof("Launching lagrande on %s, PID %s", hostname, stringPid)
	if dryRun {
		log.Info("\tDRY-RUN: no metrics will actually be sent")
	}
	log.Info("Configuration:")
	if !dryRun {
		log.Infof("\tEndpoint:")
		log.Infof("\t\tAddress: %s", endpoint)
		log.Infof("\t\tProtocol: %s", protocol)
		log.Infof("\t\tFormat: %s", format)
	}
	log.Infof("\tWorkers")
	log.Infof("\t\tCount: %d", workersCount)
	log.Infof("\t\tStart interval: %s", workersInterval)
	if !dryRun {
		log.Infof("\t\tSend interval: %s", interval)
	}
	log.Infof("\tEach worker will generate %d time series:", len(generatorsArr))
	for _, gen := range generatorsArr {
		log.Infof("\t\t- %s", gen.ToString())
	}
}

func handleStats(statsChan <-chan emissionStat) {
	accumulation := statsPrintToPushRatio * workersCount

	log.Infof("The stats print to push ratio is %d, so we'll accumulate %d data before printing.\n", statsPrintToPushRatio, accumulation)
	for { // Keep reading from channel(s) forever
		metricsSucessfullySentOverPrintWindow := int64(0)
		metricsUnsucessfullySentOverPrintWindow := int64(0)
		durationOverPrintWindow := time.Duration(0)

		for i := 0; i < accumulation; i++ { // Accumulate X stats structs to print average over all workers over the print duration
			select {
			// Need a control channel here
			case stats := <-statsChan:
				log.Tracef("Received %d/%d stats data\n", i, accumulation)
				metricsSucessfullySentOverPrintWindow += stats.successfullySent
				metricsUnsucessfullySentOverPrintWindow += stats.unsuccessfullySent
				durationOverPrintWindow += stats.duration
			}
		}

		averageSuccessfulMPS := float64(metricsSucessfullySentOverPrintWindow) / durationOverPrintWindow.Seconds()
		successRatio := float64(metricsSucessfullySentOverPrintWindow) / float64(metricsSucessfullySentOverPrintWindow+metricsUnsucessfullySentOverPrintWindow) * 100

		// Accumulaton done, we can print an average
		// TODO use logger
		log.Infof("%d workers successfully sent an average of %.3f metrics per second. A total of %s metrics were successfully sent out of %s generated. Success sent ratio if %6.2f%%\n", workersCount, averageSuccessfulMPS, humanReadableNumber(metricsSucessfullySentOverPrintWindow), humanReadableNumber(metricsSucessfullySentOverPrintWindow+metricsUnsucessfullySentOverPrintWindow), successRatio)
		// <Worker count>, <avg succ mps>, <total succ>, <total metrics>, <succ %>
		log.Infof("MRS: %d,%.3f,%s,%s,%6.2f\n", workersCount, averageSuccessfulMPS, humanReadableNumber(metricsSucessfullySentOverPrintWindow), humanReadableNumber(metricsSucessfullySentOverPrintWindow+metricsUnsucessfullySentOverPrintWindow), successRatio)
	}
}

func spawnWorker(id int, statsChan chan<- emissionStat) {
	workerFullname := fmt.Sprintf("worker-%s-%d", stringPid, id)

	workerMetricNamespacePrefix := &metricNamespacePrefix
	workerMetricNamespacePrefix = replaceOnlyIfRequired(workerMetricNamespacePrefix, "WORKERNUM", strconv.Itoa(id))
	workerMetricNamespacePrefix = replaceOnlyIfRequired(workerMetricNamespacePrefix, "WORKERFULLNAME", workerFullname)

	workerMetricNamespaceSuffix := &metricNamespaceSuffix
	workerMetricNamespaceSuffix = replaceOnlyIfRequired(workerMetricNamespaceSuffix, "WORKERNUM", strconv.Itoa(id))
	workerMetricNamespaceSuffix = replaceOnlyIfRequired(workerMetricNamespaceSuffix, "WORKERFULLNAME", workerFullname)

	var workerTags *string
	workerTags = replaceOnlyIfRequired(&workersTags, "WORKERNUM", strconv.Itoa(id))
	workerTags = replaceOnlyIfRequired(workerTags, "WORKERFULLNAME", workerFullname)

	var workerPublisher publisher.Publisher

	switch format {
	case "atlas":
		switch protocol {
		case "http":
			workerPublisher = publisher.NewHttpPublisher(endpoint)
		case "dry-run":
			workerPublisher = publisher.NewLogPublisher("na")
		}
	case "carbon":
		switch protocol {
		case "tcp":
			workerPublisher = publisher.NewTcpPublisher(endpoint)
		case "dry-run":
			workerPublisher = publisher.NewLogPublisher("na")
		}
	case "carbon-pickle":
		log.Fatal("Not supported yet")
		switch protocol {
		case "tcp":
			workerPublisher = publisher.NewTcpPublisher(endpoint)
		case "dry-run":
			workerPublisher = publisher.NewLogPublisher("na")
		}
	case "influxdb":
		switch protocol {
		case "http":
			workerPublisher = publisher.NewHttpPublisher(endpoint)
		case "dry-run":
			workerPublisher = publisher.NewLogPublisher("na")
		}
	case "m3db":
		switch protocol {
		case "http":
			workerPublisher = publisher.NewHttpPublisher(endpoint)
		case "dry-run":
			workerPublisher = publisher.NewLogPublisher("na")
		}
	case "timescale":
		switch protocol {
		case "timescale":
			workerPublisher = publisher.NewTimescalePublisher(endpoint)
		case "dry-run":
			workerPublisher = publisher.NewLogPublisher("na")
		}
	}

	workerGeneratorsArr := cloneDefaultGenerators(workerMetricNamespacePrefix, workerMetricNamespaceSuffix, workerTags)

	var metricsSucessfullyTotal int64
	var metricsUnsucessfullyTotal int64
	var metricsSucessfullyStats int64
	var metricsUnsucessfullyStats int64
	previousStatsTimestamp := time.Now()
	metricTicker := time.Tick(intervalDuration)

	for {
		select {
		// Need a control channel here
		case <-metricTicker:
			var metricArr []*metric.Metric
			// TODO replace
			metricArr = make([]*metric.Metric, len(workerGeneratorsArr), len(workerGeneratorsArr))

			for i, gen := range workerGeneratorsArr {
				metricArr[i] = gen.GenerateMetric()
			}

			var publishErr error
			formattedMetric := localFormatter.FormatData(&metricArr)
			if format == "timescale" {
				publishErr = workerPublisher.PublishMetrics(&metricArr)
			} else {
				publishErr = workerPublisher.PublishBytes(formattedMetric)
			}

			if publishErr != nil {
				metricsUnsucessfullyStats++
			} else {
				metricsSucessfullyStats++
			}

			// Push stats every XXXXms
			// TODO: Optimization: for really low sending interval, compute this every X ticks
			if previousStatsTimestamp.Add(statsPushInterval).Before(time.Now()) {
				metricsSucessfullyTotal += metricsSucessfullyStats
				metricsUnsucessfullyTotal += metricsUnsucessfullyStats

				newStatsTimestamp := time.Now()

				select {
				case statsChan <- emissionStat{workerNum: workersCount, successfullySent: metricsSucessfullyStats, unsuccessfullySent: metricsUnsucessfullyStats, duration: newStatsTimestamp.Sub(previousStatsTimestamp)}:
				default:
					log.Error("Channel full, discarding stats")
				}

				metricsSucessfullyStats = 0
				metricsUnsucessfullyStats = 0
				previousStatsTimestamp = newStatsTimestamp
			}
		}
	}
}

func cloneDefaultGenerators(workerMetricNamespacePrefix *string, workerMetricNamespaceSuffix *string, workerTags *string) []generator.Generator {
	var workerGeneratorsArr []generator.Generator
	workerGeneratorsArr = make([]generator.Generator, len(generatorsArr), len(generatorsArr))
	for i, gen := range generatorsArr {
		metricName := fmt.Sprintf("%s%s%s", *workerMetricNamespacePrefix, gen.GetName(), *workerMetricNamespaceSuffix)

		workerMetricTags := replaceOnlyIfRequired(workerTags, "METRICNAME", metricName)
		workerGeneratorsArr[i] = gen.Clone(metricName, localFormatter.FormatTags(workerMetricTags))
	}

	return workerGeneratorsArr
}

func replaceOnlyIfRequired(source *string, old string, new string) *string {
	if strings.Contains(*source, old) {
		s := strings.ReplaceAll(*source, old, new)
		return &s
	}
	// else:
	return source
}

func humanReadableNumber(number int64) string {
	if number >= 1000000000000 {
		return fmt.Sprintf("%.0fT", math.Round(float64(number)/1000000000000))
	} else if number >= 1000000000 {
		return fmt.Sprintf("%.0fG", math.Round(float64(number)/1000000000))
	} else if number >= 1000000 {
		return fmt.Sprintf("%.0fM", math.Round(float64(number)/1000000))
	} else if number >= 1000 {
		return fmt.Sprintf("%.0fK", math.Round(float64(number)/1000))
	} else {
		return fmt.Sprintf("%d", number)
	}
}
