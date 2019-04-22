package main

import (
	"time"

	"github.com/kukinsula/monitor/monitor/metric"
	"github.com/kukinsula/monitor/mq"
)

type Monitoring struct {
	config *metric.Config
	cpu    *metric.CPU
	mem    *metric.Memory
	net    *metric.Network
}

func NewMonitoring(config *metric.Config) *Monitoring {
	return &Monitoring{
		config: config,
		cpu:    metric.NewCPU(config),
		mem:    metric.NewMemory(config),
		net:    metric.NewNetwork(config),
	}
}

func (m *Monitoring) Start(channel chan *mq.Metrics) error {
	var err error

	for err == nil {
		metrics := &mq.Metrics{}

		err = m.cpu.Update()
		if err != nil {
			break
		}

		metrics.CPU = m.cpu.Public()

		err = m.mem.Update()
		if err != nil {
			break
		}

		metrics.MEM = m.mem.Public()

		err = m.net.Update()
		if err != nil {
			break
		}

		metrics.NET = m.net.Public()

		channel <- metrics

		time.Sleep(time.Second)
	}

	return err
}
