package models

import "time"

type Event struct {
	Date        time.Time `json:"date"  awsmTable:"Posted"`
	Summary     string    `json:"summary" awsmTable:"Summary"`
	ServiceName string    `json:"serviceName" awsmTable:"Service Name"`
	Details     string    `json:"details"`
	Service     string    `json:"service"`
	Archive     bool      `json:"archive"`
}
