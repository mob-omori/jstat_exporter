package main

import (
	"flag"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"bufio"

	"sync"

	"github.com/go-errors/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "jstat"
)

var (
	listenAddress = flag.String("web.listen-address", ":9010", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	jstatPath     = flag.String("jstat.path", "/usr/bin/jstat", "jstat path")
	jpsPath       = flag.String("jps.path", "/usr/bin/jps", "jps path")
	target        = flag.String("target", "", "Target name of jps.")
	interval      = flag.String("interval", "1000", "Interval of jps.")
)

var mu = sync.RWMutex{}
var latestJstat = make(map[string]string)

type Exporter struct {
	newMax     prometheus.Gauge
	newCommit  prometheus.Gauge
	oldMax     prometheus.Gauge
	oldCommit  prometheus.Gauge
	metaMax    prometheus.Gauge
	metaCommit prometheus.Gauge
	metaUsed   prometheus.Gauge
	oldUsed    prometheus.Gauge
	sv0Used    prometheus.Gauge
	sv1Used    prometheus.Gauge
	edenUsed   prometheus.Gauge
	fgcTimes   prometheus.Gauge
	fgcSec     prometheus.Gauge
}

func NewExporter() *Exporter {
	return &Exporter{
		newMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "newMax",
			Help:      "newMax",
		}),
		newCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "newCommit",
			Help:      "newCommit",
		}),
		oldMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldMax",
			Help:      "oldMax",
		}),
		oldCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldCommit",
			Help:      "oldCommit",
		}),
		metaMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaMax",
			Help:      "metaMax",
		}),
		metaCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaCommit",
			Help:      "metaCommit",
		}),
		metaUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaUsed",
			Help:      "metaUsed",
		}),
		oldUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldUsed",
			Help:      "oldUsed",
		}),
		sv0Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv0Used",
			Help:      "sv0Used",
		}),
		sv1Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv1Used",
			Help:      "sv1Used",
		}),
		edenUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "edenUsed",
			Help:      "edenUsed",
		}),
		fgcTimes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fgcTimes",
			Help:      "fgcTimes",
		}),
		fgcSec: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fgcSec",
			Help:      "fgcSec",
		}),
	}
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.newMax.Describe(ch)
	e.newCommit.Describe(ch)
	e.oldMax.Describe(ch)
	e.oldCommit.Describe(ch)
	e.metaMax.Describe(ch)
	e.metaCommit.Describe(ch)
	e.metaUsed.Describe(ch)
	e.oldUsed.Describe(ch)
	e.sv0Used.Describe(ch)
	e.sv1Used.Describe(ch)
	e.edenUsed.Describe(ch)
	e.fgcTimes.Describe(ch)
	e.fgcSec.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.JstatGccapacity(ch)
	e.JstatGcold(ch)
	e.JstatGcnew(ch)
	e.JstatGc(ch)
}

func (e *Exporter) JstatGccapacity(ch chan<- prometheus.Metric) {
	mu.RLock()
	defer mu.RUnlock()

	line, ok := latestJstat["-gccapacity"]

	if ok && line != "" {
		parts := strings.Fields(line)
		newMax, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.newMax.Set(newMax)
		e.newMax.Collect(ch)
		newCommit, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.newCommit.Set(newCommit)
		e.newCommit.Collect(ch)
		oldMax, err := strconv.ParseFloat(parts[7], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.oldMax.Set(oldMax)
		e.oldMax.Collect(ch)
		oldCommit, err := strconv.ParseFloat(parts[8], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.oldCommit.Set(oldCommit)
		e.oldCommit.Collect(ch)
		metaMax, err := strconv.ParseFloat(parts[11], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.metaMax.Set(metaMax)
		e.metaMax.Collect(ch)
		metaCommit, err := strconv.ParseFloat(parts[12], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.metaCommit.Set(metaCommit)
		e.metaCommit.Collect(ch)
	}
}

func (e *Exporter) JstatGcold(ch chan<- prometheus.Metric) {
	mu.RLock()
	defer mu.RUnlock()

	line, ok := latestJstat["-gcold"]

	if ok && line != "" {
		parts := strings.Fields(line)
		metaUsed, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.metaUsed.Set(metaUsed) // MU: Metaspace utilization (kB).
		e.metaUsed.Collect(ch)
		oldUsed, err := strconv.ParseFloat(parts[5], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.oldUsed.Set(oldUsed) // OU: Old space utilization (kB).
		e.oldUsed.Collect(ch)
	}
}

func (e *Exporter) JstatGcnew(ch chan<- prometheus.Metric) {
	mu.RLock()
	defer mu.RUnlock()

	line, ok := latestJstat["-gcnew"]

	if ok && line != "" {
		parts := strings.Fields(line)
		sv0Used, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.sv0Used.Set(sv0Used)
		e.sv0Used.Collect(ch)
		sv1Used, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.sv1Used.Set(sv1Used)
		e.sv1Used.Collect(ch)
		edenUsed, err := strconv.ParseFloat(parts[8], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.edenUsed.Set(edenUsed)
		e.edenUsed.Collect(ch)
	}
}

func (e *Exporter) JstatGc(ch chan<- prometheus.Metric) {
	mu.RLock()
	defer mu.RUnlock()

	line, ok := latestJstat["-gc"]

	if ok && line != "" {
		parts := strings.Fields(line)
		fgcTimes, err := strconv.ParseFloat(parts[14], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.fgcTimes.Set(fgcTimes)
		e.fgcTimes.Collect(ch)
		fgcSec, err := strconv.ParseFloat(parts[15], 64)
		if err != nil {
			log.Fatal(err)
		}
		e.fgcSec.Set(fgcSec)
		e.fgcSec.Collect(ch)
	}
}

func RunJstatGccapacity(jstatPath, target, interval string) {
	runCommand(jstatPath, "-gccapacity", target, interval)
}

func RunJstatGcold(jstatPath, target, interval string) {
	runCommand(jstatPath, "-gcold", target, interval)
}

func RunJstatGcnew(jstatPath, target, interval string) {
	runCommand(jstatPath, "-gcnew", target, interval)
}

func RunJstatGc(jstatPath, target, interval string) {
	runCommand(jstatPath, "-gc", target, interval)
}

func runCommand(jstatPath, command, target, interval string) {

	for {
		pid, err := Jps(target)
		if err != nil {
			time.Sleep(60 * time.Second)
			continue
		}

		cmd := exec.Command(jstatPath, command, pid, interval)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Error(err)
			continue
		}

		cmd.Start()

		scanner := bufio.NewScanner(stdout)

		first := true
		for scanner.Scan() {
			line := scanner.Text()
			if first {
				first = false
				continue
			}
			// log.Println("put... ")
			mu.Lock()
			latestJstat[command] = line
			mu.Unlock()
		}

		killProcess(cmd)

		// end jstat
		log.Printf("Finished jstat... restart")
	}
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			log.Error("Error killing jstat process: %v", err)
		}
	}
}

func Jps(name string) (string, error) {
	cmd := exec.Command(*jpsPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("Error get stdout pipe: %v", err)
		return "", err
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	pid := ""
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, " ")

		if len(items) == 2 {
			if items[1] == "Jps" || items[1] == "Jstat" {
				continue
			}

			if name != "" {
				if items[1] == name {
					pid = items[0]
					break
				}
			} else {
				pid = items[0]
				break
			}
		}
	}
	cmd.Wait()

	if len(pid) == 0 {
		log.Error("No target process: %v", name)
		return "", errors.New("No target process: " + name)
	}

	return pid, nil
}

func main() {
	flag.Parse()

	go RunJstatGc(*jstatPath, *target, *interval)
	go RunJstatGcnew(*jstatPath, *target, *interval)
	go RunJstatGcold(*jstatPath, *target, *interval)
	go RunJstatGccapacity(*jstatPath, *target, *interval)

	exporter := NewExporter()
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>jstat Exporter</title></head>
		<body>
		<h1>jstat Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

}
