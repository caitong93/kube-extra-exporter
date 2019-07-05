package metrics

import (
	"github.com/caicloud/nirvana/log"
	"github.com/caitong93/kube-extra-exporter/pkg/info"
	"github.com/prometheus/client_golang/prometheus"
)

type podMetric struct {
	name        string
	help        string
	valueType   prometheus.ValueType
	extraLabels []string
	getValues   func(s *info.Stats) metricValues
}

type metricValues []metricValue

// metricValue describes a single metric value for a given set of label values
// within a parent podMetric.
type metricValue struct {
	value  float64
	labels []string
}

type infoProvider interface {
	ListStats() ([]*info.Stats, error)
}

// PrometheusCollector implements prometheus.Collector.
type PrometheusCollector struct {
	infoProvider infoProvider
	errors       prometheus.Gauge
	podMetrics   []podMetric
}

func NewPrometheusCollector(i infoProvider) *PrometheusCollector {
	return &PrometheusCollector{
		infoProvider: i,
		errors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "container",
			Name:      "scrape_error",
			Help:      "1 if there was an error while getting container metrics, 0 otherwise",
		}),
		podMetrics: []podMetric{
			{
				name:        "pod_tcp_connections",
				help:        "tcp(include tcp6) connections usage statistic for pod",
				valueType:   prometheus.GaugeValue,
				extraLabels: []string{"tcp_state", "proto"},
				getValues: func(s *info.Stats) metricValues {
					return metricValues{
						{
							value:  float64(s.Network.Tcp.Established),
							labels: []string{"established", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.SynSent),
							labels: []string{"synsent", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.SynRecv),
							labels: []string{"synrecv", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.FinWait1),
							labels: []string{"finwait1", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.FinWait2),
							labels: []string{"finwait2", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.TimeWait),
							labels: []string{"timewait", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.Close),
							labels: []string{"close", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.CloseWait),
							labels: []string{"closewait", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.LastAck),
							labels: []string{"lastack", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.Listen),
							labels: []string{"listen", "tcp"},
						},
						{
							value:  float64(s.Network.Tcp.Closing),
							labels: []string{"closing", "tcp"},
						},
						// Tcp6 stats
						{
							value:  float64(s.Network.Tcp6.Established),
							labels: []string{"established", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.SynSent),
							labels: []string{"synsent", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.SynRecv),
							labels: []string{"synrecv", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.FinWait1),
							labels: []string{"finwait1", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.FinWait2),
							labels: []string{"finwait2", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.TimeWait),
							labels: []string{"timewait", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.Close),
							labels: []string{"close", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.CloseWait),
							labels: []string{"closewait", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.LastAck),
							labels: []string{"lastack", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.Listen),
							labels: []string{"listen", "tcp6"},
						},
						{
							value:  float64(s.Network.Tcp6.Closing),
							labels: []string{"closing", "tcp6"},
						},
					}
				},
			},
		},
	}
}

// Collect fetches the stats from all containers and delivers them as
// Prometheus metrics. It implements prometheus.PrometheusCollector.
func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	c.errors.Set(0)
	c.collectPodsInfo(ch)
	c.errors.Collect(ch)
}

func defaultPodLabels(i *info.Stats) map[string]string {
	labels := map[string]string{}
	labels["pod"] = i.PodName
	labels["namespace"] = i.Namespace
	return labels
}

func (c *PrometheusCollector) collectPodsInfo(ch chan<- prometheus.Metric) {
	infos, err := c.infoProvider.ListStats()
	if err != nil {
		c.errors.Set(1)
		log.Errorf("err get pod infos: %v", err)
		return
	}

	for _, info := range infos {
		podLabels := defaultPodLabels(info)
		labels := []string{}
		values := []string{}
		for lk, lv := range podLabels {
			labels = append(labels, lk)
			values = append(values, lv)
		}

		for _, metric := range c.podMetrics {

			desc := metric.desc(labels)
			for _, v := range metric.getValues(info) {
				ch <- prometheus.MustNewConstMetric(desc, metric.valueType, v.value, append(values, v.labels...)...)
			}
		}
	}
}

// Describe describes all the metrics ever exported by cadvisor. It
// implements prometheus.PrometheusCollector.
func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	c.errors.Describe(ch)
	for _, m := range c.podMetrics {
		ch <- m.desc([]string{})
	}
}

func (m *podMetric) desc(baseLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(m.name, m.help, append(baseLabels, m.extraLabels...), nil)
}
