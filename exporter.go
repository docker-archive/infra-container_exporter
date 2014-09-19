package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/docker/libcontainer/cgroups/fs"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "container" // For Prometheus metrics.
)

var (
	pagingCounters = []string{
		"pgpgin", "pgpgout", "pgfault", "pgmajfault",
		"total_pgpgin", "total_pgpgout", "total_pgfault", "total_pgmajfault",
	}
)

func isMemoryPagingCounter(t string) bool {
	for _, counter := range pagingCounters {
		if t == counter {
			return true
		}
	}
	return false
}

// Exporter collects container stats and exports them using
// the prometheus metrics package.
type Exporter struct {
	mutex   sync.RWMutex
	manager Manager

	errors                       *prometheus.CounterVec
	lastSeen                     *prometheus.CounterVec
	cpuUsageSeconds              *prometheus.CounterVec
	cpuUsageSecondsPerCPU        *prometheus.CounterVec
	cpuThrottledPeriods          *prometheus.CounterVec
	cpuThrottledTime             *prometheus.CounterVec
	memoryUsageBytes             *prometheus.GaugeVec
	memoryMaxUsageBytes          *prometheus.GaugeVec
	memoryFailures               *prometheus.CounterVec
	memoryStats                  *prometheus.GaugeVec
	memoryPaging                 *prometheus.CounterVec
	blkioIoServiceBytesRecursive *prometheus.CounterVec
	blkioIoServicedRecursive     *prometheus.CounterVec
	blkioIoQueuedRecursive       *prometheus.GaugeVec
	blkioSectorsRecursive        *prometheus.CounterVec
}

// NewExporter returns an initialized Exporter.Vec
func NewExporter(manager Manager) *Exporter {
	return &Exporter{
		manager: manager,
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Errors while exporting container metrics.",
		},
			[]string{"component"},
		),
		lastSeen: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "last_seen",
			Help:      "Last time a container was seen by the exporter",
		},
			[]string{"name", "id"},
		),
		cpuUsageSeconds: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cpu_usage_seconds_total",
			Help:      "Total seconds of cpu time consumed.",
		},
			[]string{"name", "id", "type"},
		),
		cpuUsageSecondsPerCPU: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cpu_usage_per_cpu_seconds_total",
			Help:      "Total seconds of cpu time consumed per cpu.",
		},
			[]string{"name", "id", "cpu"},
		),

		cpuThrottledPeriods: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cpu_throttled_periods_total",
			Help:      "Number of periods with throttling.",
		},
			[]string{"name", "id", "state"},
		),
		cpuThrottledTime: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "cpu_throttled_time_seconds_total",
			Help:      "Aggregate time the container was throttled for in seconds.",
		},
			[]string{"name", "id"},
		),
		memoryUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes.",
		},
			[]string{"name", "id"},
		),
		memoryMaxUsageBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_max_usage_bytes",
			Help:      "Maximum memory usage ever recorded in bytes.",
		},
			[]string{"name", "id"},
		),
		memoryFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "memory_failures_total",
			Help:      "Number of times memory usage hits limits.",
		},
			[]string{"name", "id"},
		),
		// Since libcontainer exports this only as a raw map, we just expose those
		// metrics like this. This may change (and break) in the future.
		memoryStats: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_stats",
			Help:      "Stats from cgroup/memory/memory.stat.",
		},
			[]string{"name", "id", "type"},
		),
		memoryPaging: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "memory_paging_total",
			Help:      "Paging events from cgroup/memory/memory.stat.",
		},
			[]string{"name", "id", "type"},
		),
		blkioIoServiceBytesRecursive: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "blkio_io_service_bytes_recursive_total",
			Help:      "Number of bytes transferred to/from the disk by the cgroup.",
		},
			[]string{"name", "id", "device", "op"},
		),
		blkioIoServicedRecursive: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "blkio_io_serviced_recursive_total",
			Help:      "Number of IOs completed to/from the disk by the cgroup.",
		},
			[]string{"name", "id", "device", "op"},
		),
		blkioIoQueuedRecursive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "blkio_io_queued_recursive",
			Help:      "Number of requests currently queued up for the cgroup.",
		},
			[]string{"name", "id", "device", "op"},
		),
		blkioSectorsRecursive: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "blkio_sectors_recursive_total",
			Help:      "Number of sectors transferred to/from disk by the cgroup.",
		},
			[]string{"name", "id", "device"},
		),
	}
}

// Describe describes all the metrics ever exported. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.errors.Describe(ch)
	e.lastSeen.Describe(ch)
	e.cpuUsageSeconds.Describe(ch)
	e.cpuUsageSecondsPerCPU.Describe(ch)
	e.cpuThrottledPeriods.Describe(ch)
	e.cpuThrottledTime.Describe(ch)
	e.memoryUsageBytes.Describe(ch)
	e.memoryMaxUsageBytes.Describe(ch)
	e.memoryFailures.Describe(ch)
	e.memoryStats.Describe(ch)
	e.memoryPaging.Describe(ch)
	e.blkioIoServiceBytesRecursive.Describe(ch)
	e.blkioIoServicedRecursive.Describe(ch)
	e.blkioIoQueuedRecursive.Describe(ch)
	e.blkioSectorsRecursive.Describe(ch)
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	containers, err := e.manager.Containers()
	if err != nil {
		e.errors.WithLabelValues("list").Inc()
		return err
	}
	for _, container := range containers {
		stats, err := fs.GetStats(container.Cgroups)
		if err != nil {
			e.errors.WithLabelValues("stats").Inc()
			return err
		}
		name := container.Name
		id := container.ID

		// Last seen
		e.lastSeen.WithLabelValues(name, id).Set(float64(time.Now().Unix()))

		// CPU stats
		// - Usage
		for i, value := range stats.CpuStats.CpuUsage.PercpuUsage {
			e.cpuUsageSecondsPerCPU.WithLabelValues(name, id, fmt.Sprintf("cpu%02d", i)).Set(float64(value) / float64(time.Second))
		}

		e.cpuUsageSeconds.WithLabelValues(name, id, "kernel").Set(float64(stats.CpuStats.CpuUsage.UsageInKernelmode) / float64(time.Second))
		e.cpuUsageSeconds.WithLabelValues(name, id, "user").Set(float64(stats.CpuStats.CpuUsage.UsageInUsermode) / float64(time.Second))

		// - Throttling
		e.cpuThrottledPeriods.WithLabelValues(name, id, "total").Set(float64(stats.CpuStats.ThrottlingData.Periods))
		e.cpuThrottledPeriods.WithLabelValues(name, id, "throttled").Set(float64(stats.CpuStats.ThrottlingData.ThrottledPeriods))
		e.cpuThrottledTime.WithLabelValues(name, id).Set(float64(stats.CpuStats.ThrottlingData.ThrottledTime) / float64(time.Second))

		// Memory stats
		e.memoryUsageBytes.WithLabelValues(name, id).Set(float64(stats.MemoryStats.Usage))
		e.memoryMaxUsageBytes.WithLabelValues(name, id).Set(float64(stats.MemoryStats.MaxUsage))

		for t, value := range stats.MemoryStats.Stats {
			if isMemoryPagingCounter(t) {
				e.memoryPaging.WithLabelValues(name, id, t).Set(float64(value))
			} else {
				e.memoryStats.WithLabelValues(name, id, t).Set(float64(value))
			}
		}
		e.memoryFailures.WithLabelValues(name, id).Set(float64(stats.MemoryStats.Failcnt))

		// BlkioStats
		devMap, err := newDeviceMap(procDiskStats)
		if err != nil {
			return err
		}
		for _, stat := range stats.BlkioStats.IoServiceBytesRecursive {
			e.blkioIoServiceBytesRecursive.WithLabelValues(name, id, devMap.name(stat.Major, stat.Minor), stat.Op).Set(float64(stat.Value))
		}
		for _, stat := range stats.BlkioStats.IoServicedRecursive {
			e.blkioIoServicedRecursive.WithLabelValues(name, id, devMap.name(stat.Major, stat.Minor), stat.Op).Set(float64(stat.Value))
		}
		for _, stat := range stats.BlkioStats.IoQueuedRecursive {
			e.blkioIoQueuedRecursive.WithLabelValues(name, id, devMap.name(stat.Major, stat.Minor), stat.Op).Set(float64(stat.Value))
		}
		for _, stat := range stats.BlkioStats.SectorsRecursive {
			e.blkioSectorsRecursive.WithLabelValues(name, id, devMap.name(stat.Major, stat.Minor)).Set(float64(stat.Value))
		}
	}
	return nil
}

// Collect fetches the stats from container manager and system and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	e.lastSeen.Reset()
	e.cpuUsageSeconds.Reset()
	e.cpuUsageSecondsPerCPU.Reset()
	e.cpuThrottledPeriods.Reset()
	e.cpuThrottledTime.Reset()
	e.memoryUsageBytes.Reset()
	e.memoryMaxUsageBytes.Reset()
	e.memoryFailures.Reset()
	e.memoryStats.Reset()
	e.memoryPaging.Reset()
	e.blkioIoServiceBytesRecursive.Reset()
	e.blkioIoServicedRecursive.Reset()
	e.blkioIoQueuedRecursive.Reset()
	e.blkioSectorsRecursive.Reset()
	if err := e.collect(ch); err != nil {
		log.Printf("Error reading container stats: %s", err)
		e.errors.WithLabelValues("collect").Inc()
	}
	e.errors.Collect(ch)
	e.lastSeen.Collect(ch)
	e.cpuUsageSeconds.Collect(ch)
	e.cpuUsageSecondsPerCPU.Collect(ch)
	e.cpuThrottledPeriods.Collect(ch)
	e.cpuThrottledTime.Collect(ch)
	e.memoryUsageBytes.Collect(ch)
	e.memoryMaxUsageBytes.Collect(ch)
	e.memoryFailures.Collect(ch)
	e.memoryStats.Collect(ch)
	e.memoryPaging.Collect(ch)
	e.blkioIoServiceBytesRecursive.Collect(ch)
	e.blkioIoServicedRecursive.Collect(ch)
	e.blkioIoQueuedRecursive.Collect(ch)
	e.blkioSectorsRecursive.Collect(ch)
}
