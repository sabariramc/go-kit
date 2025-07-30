// Package correlation enhances the context of requests with correlation and user identity.
package correlation

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// EventCorrelation defines a context object for correlation.
type EventCorrelation struct {
	CorrelationID string
	ScenarioID    string
	SessionID     string
	ScenarioName  string
}

func (c *EventCorrelation) MarshalJSON() ([]byte, error) {
	sb := &strings.Builder{}
	sb.Grow(len(c.CorrelationID) + 200)
	sb.Write([]byte("{"))
	addComma := false
	valueList := make([][2]string, 0, 5)
	if c.CorrelationID != "" {
		valueList = append(valueList, [2]string{"correlationID", c.CorrelationID})
	}
	if c.ScenarioID != "" {
		valueList = append(valueList, [2]string{"scenarioID", c.ScenarioID})
	}
	if c.SessionID != "" {
		valueList = append(valueList, [2]string{"sessionID", c.SessionID})
	}
	if c.ScenarioName != "" {
		valueList = append(valueList, [2]string{"scenarioName", c.ScenarioName})
	}
	for _, v := range valueList {
		if addComma {
			sb.WriteString(",")
		}
		sb.WriteString("\"" + v[0] + "\":\"" + v[1] + "\"")
		addComma = true
	}
	sb.WriteString("}")
	return []byte(sb.String()), nil
}

const (
	CorrelationIDHeader = "X-Correlation-ID"
	ScenarioIDHeader    = "X-Scenario-ID"
	SessionIDHeader     = "X-Session-ID"
	ScenarioNameHeader  = "X-Scenario-Name"
)

func (c *EventCorrelation) GetHeader() map[string]string {
	headers := make(map[string]string, 10)
	if c.CorrelationID != "" {
		headers[CorrelationIDHeader] = c.CorrelationID
	}
	if c.ScenarioID != "" {
		headers[ScenarioIDHeader] = c.ScenarioID
	}
	if c.SessionID != "" {
		headers[SessionIDHeader] = c.SessionID
	}
	if c.ScenarioName != "" {
		headers[ScenarioNameHeader] = c.ScenarioName
	}
	return headers
}

func NewCorrelationParam(serviceName string) *EventCorrelation {
	return &EventCorrelation{
		CorrelationID: fmt.Sprintf("%v-%v", serviceName, uuid.New().String()),
	}
}
