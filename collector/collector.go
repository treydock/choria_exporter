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

package collector

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "mcollective"
)

var (
	mcoPath         = kingpin.Flag("path.mco", "Path to mco").Default("/opt/puppetlabs/bin/mco").String()
	collectorState  = make(map[string]*bool)
	factories       = make(map[string]func(logger log.Logger, host string) Collector)
	execCommand     = exec.Command
	collectDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil)
	collectError = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collect_error"),
		"Indicates if error has occurred during collection",
		[]string{"collector"}, nil)
)

type McollectiveCollector struct {
	Collectors map[string]Collector
}

type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
}

func registerCollector(collector string, isDefaultEnabled bool, factory func(logger log.Logger, host string) Collector) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}
	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)
	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Bool()
	collectorState[collector] = flag
	factories[collector] = factory
}

func NewMcollectiveCollector(logger log.Logger, host string) *McollectiveCollector {
	if !fileExists(*mcoPath) {
		level.Error(logger).Log("error", fmt.Sprintf("Path %s for mco does not exist", *mcoPath))
		os.Exit(1)
	}
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		var collector Collector
		if *enabled {
			collector = factories[key](logger, host)
			collectors[key] = collector
		}
	}
	return &McollectiveCollector{Collectors: collectors}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
