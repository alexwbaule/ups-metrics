package graylog

import (
	"fmt"
	"github.com/alexwbaule/ups-metrics/internal/application"
	"github.com/alexwbaule/ups-metrics/internal/application/logger"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
	"time"
)

type Gelf struct {
	Address  string
	gelf     *gelf.Writer
	Hostname string
	log      *logger.Logger
}

func NewGelf(l *application.Application) *Gelf {
	cf := l.Config.GetGelfConfig()

	g, err := gelf.NewWriter(fmt.Sprintf("%s:%s", cf.Address, cf.Port))
	if err != nil {
		l.Log.Infof("Error creating Gelf Writer: %s", err.Error())
	}
	return &Gelf{
		Address:  fmt.Sprintf("%s:%s", cf.Address, cf.Port),
		gelf:     g,
		Hostname: l.Config.GetDeviceAddress(),
		log:      l.Log,
	}

}

func (m *Gelf) LogNotifications(not device.Notification) {
	var dt time.Time
	var full string
	extraMessage := map[string]interface{}{
		"application_name": "ups-metrics",
		"id":               not.ID,
		"message":          not.Message,
		"date":             not.Date,
	}

	parse, err := time.ParseInLocation("02/01/2006 15:04:05", not.Date, time.Local)
	if err != nil {
		m.log.Infof("Error parsing date [%s]: %s", not.Date, err.Error())
		dt = time.Now()
	} else {
		dt = parse
	}

	full = fmt.Sprintf("Notification %d on %s with %s", not.ID, not.Date, not.Message)

	msg := &gelf.Message{
		Version:  "1.1",
		Host:     m.Hostname,
		Short:    full,
		TimeUnix: float64(dt.Unix()),
		Level:    6,
		Facility: "ups-metrics",
		Extra:    extraMessage,
	}
	err = m.gelf.WriteMessage(msg)
	if err != nil {
		m.log.Infof("Error writing message: %s", err.Error())
	}
	m.log.Infof("Sended: %s", full)
}

func (m *Gelf) Disconnect() {
	_ = m.gelf.Close()
}
