// Copyright 2020 Trey Dockendorf
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collectors

import (
	"bytes"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configPingTimeout = kingpin.Flag("collector.ping.timeout", "Timeout for mco ping").Default("1").String()
)

type PingMetric struct {
	Status int
	Time   float64
}

type PingCollector struct {
	host   string
	Status *prometheus.Desc
	Time   *prometheus.Desc
}

func init() {
	registerCollector("ping", true, NewPingCollector)
}

func NewPingCollector(host string) Collector {
	return &PingCollector{
		host: host,
		Status: prometheus.NewDesc(prometheus.BuildFQName(namespace, "ping", "status"),
			"Mcollective ping status, 1=successful 0=not successful", nil, nil),
		Time: prometheus.NewDesc(prometheus.BuildFQName(namespace, "ping", "seconds"),
			"Mcollective ping time in seconds", nil, nil),
	}
}

func (c *PingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
	ch <- c.Time
}

func (c *PingCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Collecting ping metric")
	err := c.collect(ch)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 1, "ping")
	} else {
		ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, 0, "ping")
	}
}

func (c *PingCollector) collect(ch chan<- prometheus.Metric) error {
	collectTime := time.Now()
	metric, err := ping(c.host)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(c.Status, prometheus.GaugeValue, float64(metric.Status))
	ch <- prometheus.MustNewConstMetric(c.Time, prometheus.GaugeValue, metric.Time)
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "ping")
	return nil
}

func ping(host string) (PingMetric, error) {
	var metric PingMetric
	mco := *mcoPath
	timeout := *configPingTimeout
	cmd := execCommand(mco, "ping", "--timeout", timeout, "-I", host)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	re := regexp.MustCompile(`\r?\n`)
	outlog := re.ReplaceAllString(out.String(), " ")
	if err != nil {
		log.Errorf("PING: %s : %s", outlog, err)
		metric.Status = 0
	} else {
		metric.Status = 1
	}
	timePattern := regexp.MustCompile(`time=([0-9.]+) ([a-z]+)`)
	timeMatch := timePattern.FindStringSubmatch(out.String())
	if len(timeMatch) == 3 {
		time, err := strconv.ParseFloat(timeMatch[1], 64)
		if err != nil {
			log.Errorf("Error parsing time %s for %s: %s", outlog, host, err.Error())
			return metric, err
		}
		unit := timeMatch[2]
		switch unit {
		case "ms":
			time = time / 1000.0
		}
		metric.Time = time
	}
	return metric, nil
}
