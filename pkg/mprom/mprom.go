package mprom

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/encoding/prototext"
)

type tsConf struct {
	promReg *prometheus.Registry
}

type tsOpts func(*tsConf)

func New(opts ...tsOpts) *tsConf {
	t := &tsConf{
		promReg: prometheus.NewRegistry(),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// 从 Prometheus 注册器中获取指标
func (t *tsConf) GetMetrics() ([]string, error) {
	if t.promReg == nil {
		return nil, errors.New("prometheus registry is nil")
	}

	metricFamilies, err := t.promReg.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "prometheus registry gather error")
	}
	metrics := make([]string, 0)
	// 遍历所有指标族
	for _, mf := range metricFamilies {
		for _, m := range mf.GetMetric() {
			metricName := mf.GetName()
			if m.Gauge != nil {
				mapp := map[string]string{
					"__name__": metricName,
				}
				rawLabel := m.GetLabel()
				for _, l := range rawLabel {
					mapp[l.GetName()] = l.GetValue()
				}

				promMetric := prometheus.MustNewConstMetric(
					prometheus.NewDesc(metricName, "", nil, mapp),
					prometheus.GaugeValue,
					m.Gauge.GetValue(),
				)

				// 将 Metric 转换为文本格式
				var dtoMetric dto.Metric
				if err := promMetric.Write(&dtoMetric); err != nil {
					continue
				}
				metricStr := prototext.Format(&dtoMetric)
				metrics = append(metrics, metricStr)
			}
		}
	}
	return metrics, nil
}
