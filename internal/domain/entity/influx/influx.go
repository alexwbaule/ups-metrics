package influx

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
		"Nobreak":          "on_nobreak",
		"Rede Eletrica":    "on_grid",
		"Teste":            "on_test",
		"Boost":            "on_boost",
		"ByPass":           "on_bypass",
		"Potencia Elevada": "overload",
	}
	return status[code]
}
