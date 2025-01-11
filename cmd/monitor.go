package main

type Monitor interface {
	Start()
	Stop()
}

func RunMonitor(m Monitor) {
	m.Start()
	defer m.Stop()
}
