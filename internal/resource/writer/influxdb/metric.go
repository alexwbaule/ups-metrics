package influxdb

var UPSMetricName = func(code string) string {
	status := map[string]string{
		"Tensao de Entrada":   "input_voltage",
		"Tensao de Saida":     "output_voltage",
		"Nivel da Bateria":    "battery_level",
		"Potencia de Saida":   "ups_load",
		"Temperatura":         "ups_temperature",
		"Frequencia de Saida": "output_frequency",
	}
	return status[code]
}

var UPSMetricState = func(code string) string {
	status := map[string]string{
		"Carga da Bateria": "battery_status",
		"Nobreak":          "nobreak_status",
		"Rede Eletrica":    "power_from",
		"Teste":            "test",
		"Boost":            "boost",
		"ByPass":           "bypass",
		"Potencia Elevada": "overload",
	}
	return status[code]
}

var UPSMetricStateValue = func(code string, value bool) string {
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
