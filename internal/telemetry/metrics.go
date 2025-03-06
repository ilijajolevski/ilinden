// Metrics registration and collection
//
// Prometheus metrics setup:
// - Counter definitions
// - Histogram definitions
// - Gauge definitions
// - Label schemas
// - Metrics initialization

package telemetry

import (
	"fmt"
	"sync"
	"time"
)

// Metrics defines the interface for metrics collection
type Metrics interface {
	// Counter operations
	IncCounter(name string)
	IncCounterBy(name string, value int)
	
	// Gauge operations
	SetGauge(name string, value float64)
	IncGauge(name string)
	DecGauge(name string)
	
	// Histogram operations
	ObserveHistogram(name string, value float64)
	
	// Duration operations
	ObserveRequestDuration(path string, duration time.Duration)
	ObserveOriginDuration(host string, duration time.Duration)
}

// SimpleMetrics is a simple implementation of the Metrics interface
type SimpleMetrics struct {
	counters   map[string]int
	gauges     map[string]float64
	histograms map[string][]float64
	mu         sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics() Metrics {
	return &SimpleMetrics{
		counters:   make(map[string]int),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
	}
}

// IncCounter increments a counter
func (m *SimpleMetrics) IncCounter(name string) {
	m.IncCounterBy(name, 1)
}

// IncCounterBy increments a counter by a value
func (m *SimpleMetrics) IncCounterBy(name string, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.counters[name]; !exists {
		m.counters[name] = 0
	}
	
	m.counters[name] += value
}

// SetGauge sets a gauge value
func (m *SimpleMetrics) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.gauges[name] = value
}

// IncGauge increments a gauge
func (m *SimpleMetrics) IncGauge(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.gauges[name]; !exists {
		m.gauges[name] = 0
	}
	
	m.gauges[name]++
}

// DecGauge decrements a gauge
func (m *SimpleMetrics) DecGauge(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.gauges[name]; !exists {
		m.gauges[name] = 0
	}
	
	m.gauges[name]--
}

// ObserveHistogram records a histogram observation
func (m *SimpleMetrics) ObserveHistogram(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.histograms[name]; !exists {
		m.histograms[name] = make([]float64, 0)
	}
	
	m.histograms[name] = append(m.histograms[name], value)
}

// ObserveRequestDuration records the duration of a request
func (m *SimpleMetrics) ObserveRequestDuration(path string, duration time.Duration) {
	name := fmt.Sprintf("request_duration_%s", path)
	m.ObserveHistogram(name, float64(duration.Milliseconds()))
}

// ObserveOriginDuration records the duration of an origin request
func (m *SimpleMetrics) ObserveOriginDuration(host string, duration time.Duration) {
	name := fmt.Sprintf("origin_duration_%s", host)
	m.ObserveHistogram(name, float64(duration.Milliseconds()))
}

// DumpMetrics returns all metrics (for debugging)
func (m *SimpleMetrics) DumpMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := make(map[string]interface{})
	
	// Copy counters
	for k, v := range m.counters {
		metrics["counter_"+k] = v
	}
	
	// Copy gauges
	for k, v := range m.gauges {
		metrics["gauge_"+k] = v
	}
	
	// Compute histogram stats
	for k, v := range m.histograms {
		if len(v) > 0 {
			// Just store the count for simplicity
			metrics["histogram_"+k+"_count"] = len(v)
		}
	}
	
	return metrics
}