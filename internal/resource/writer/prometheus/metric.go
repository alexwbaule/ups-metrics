package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MetricStatus = map[string]*prometheus.GaugeVec{
		"Tensao de Entrada": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "input_voltage",
			Help:      "The input voltage of the UPS",
		}, []string{"host", "type", "unit"}),
		"Tensao de Saida": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "output_voltage",
			Help:      "The output voltage of the UPS",
		}, []string{"host", "type", "unit"}),
		"Nivel da Bateria": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "battery_level",
			Help:      "The battery level of the UPS",
		}, []string{"host", "type", "unit"}),
		"Potencia de Saida": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "ups_load",
			Help:      "The load of the UPS",
		}, []string{"host", "type", "unit"}),
		"Temperatura": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "ups_temperature",
			Help:      "The temperature of the UPS",
		}, []string{"host", "type", "unit"}),
		"Frequencia de Saida": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "output_frequency",
			Help:      "The output frequency of the UPS",
		}, []string{"host", "type", "unit"}),
	}

	MetricState = map[string]*prometheus.GaugeVec{
		"Carga da Bateria": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "battery_status",
			Help:      "The battery status of the UPS",
		}, []string{"host", "status"}),
		"Nobreak": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "nobreak_status",
			Help:      "The nobreak status of the UPS",
		}, []string{"host", "status"}),
		"Rede Eletrica": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "power_from",
			Help:      "The power source of the UPS",
		}, []string{"host", "status"}),
		"Teste": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "test",
			Help:      "The test status of the UPS (running or not)",
		}, []string{"host", "status"}),
		"Boost": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "boost",
			Help:      "The boost status of the UPS (running or not)",
		}, []string{"host", "status"}),
		"ByPass": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "bypass",
			Help:      "The bypass status of the UPS (active or inactive)",
		}, []string{"host", "status"}),
		"Potencia Elevada": promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ups",
			Name:      "overload",
			Help:      "The overload status of the UPS ",
		}, []string{"host", "status"}),
	}
)

var UPSMetricName = func(code string) *prometheus.GaugeVec {
	return MetricStatus[code]
}

var UPSMetricState = func(code string) *prometheus.GaugeVec {
	return MetricState[code]
}

var UPSMetricStateValue = func(code string, value bool) float64 {
	status := map[string]map[bool]float64{
		"Carga da Bateria": {
			true:  1.0,
			false: 0.0,
		},
		"Nobreak": {
			true:  0.0,
			false: 1.0,
		},
		"Rede Eletrica": {
			true:  1.0,
			false: 0.0,
		},
		"Teste": {
			true:  1.0,
			false: 0.0,
		},
		"Boost": {
			true:  1.0,
			false: 0.0,
		},
		"ByPass": {
			true:  1.0,
			false: 0.0,
		},
		"Potencia Elevada": {
			true:  1.0,
			false: 0.0,
		},
	}
	return status[code][value]
}

var UPSMetricStateLabel = func(code string, value bool) string {
	status := map[string]map[bool]string{
		"Carga da Bateria": {
			true:  "ok",
			false: "fail",
		},
		"Nobreak": {
			true:  "fail",
			false: "ok",
		},
		"Rede Eletrica": {
			true:  "grid",
			false: "battery",
		},
		"Teste": {
			true:  "on",
			false: "off",
		},
		"Boost": {
			true:  "on",
			false: "off",
		},
		"ByPass": {
			true:  "on",
			false: "off",
		},
		"Potencia Elevada": {
			true:  "true",
			false: "false",
		},
	}
	return status[code][value]
}
