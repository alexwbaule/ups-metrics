package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	status = map[string]string{
		"Tensao de Entrada":   "input_voltage",
		"Tensao de Saida":     "output_voltage",
		"Nivel da Bateria":    "battery_level",
		"Potencia de Saida":   "ups_load",
		"Temperatura":         "ups_temperature",
		"Frequencia de Saida": "output_frequency",
	}

	states = map[string]string{
		"Carga da Bateria": "battery_is_healthy",
		"Nobreak":          "nobreak_is_healthy",
		"Rede Eletrica":    "on_grid",
		"Teste":            "on_test",
		"Boost":            "on_boost",
		"ByPass":           "on_bypass",
		"Potencia Elevada": "on_high_power",
	}
)

var UPSMetricName = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ups",
	Name:      "status",
	Help:      "The status of the UPS",
}, []string{"host", "type", "unit"})

var UPSMetricStatusLabel = func(code string) string {
	return status[code]
}

var UPSMetricState = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "ups",
	Name:      "state",
	Help:      "The states of the UPS",
}, []string{"host", "state"})

var UPSMetricStateLabel = func(code string, value bool) (string, float64) {
	var v float64
	if value {
		v = 1
	}
	return states[code], v
}
