package googleanalyticsexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// Config defines the configuration for the Google Analytics Exporter.
type Config struct {
	// APIKey to report transaction to Google Analytics. If the APIKey is not set, no trace will be sent to Google Analytics.
	APISecret     string
	MeasurementID string
}

type exporter struct {
	Config *Config
	logger *zap.Logger
}

type GAMPEvent struct {
	name   string
	params map[string]interface{}
}

type GAMPRequestPayload struct {
	client_id            string
	user_id              string
	timestamp_micros     int64
	non_personalized_ads bool
	events               *[]GAMPEvent
	validationBehavior   string
}

func (e *exporter) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	logEvents, filtered := logsToGALogs(e.logger, ld)
	if len(logEvents) == 0 {
		return nil
	}

	if filtered != 0 {
		e.logger.Error("Failed ", zap.Int("num_of_events", len(logEvents)))
	}
	ts_microseconds := time.Now().UnixNano() / 1000
	mp_request := GAMPRequestPayload{
		timestamp_micros: ts_microseconds,
	}
	bodyJSON, err := json.Marshal(mp_request)
	if err != nil {
		return err
	}
	request_url := fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s", e.Config.MeasurementID, e.Config.APISecret)
	resp, err := http.Post(request_url, "application/json", bytes.NewBuffer((bodyJSON)))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Failed to send logs to Google Analytics.")
	}
	return nil
}

// returns event, valid, error
func logToGAMPEvent(log plog.LogRecord) (*GAMPEvent, bool, error) {

	log_type, valid := log.Attributes().Get("log_type")

	if !valid {
		return nil, false, nil
	}

	if log_type.Str() != "ga_event" {
		return nil, false, nil
	}

	log_params, valid := log.Attributes().Get("ga_params")

	if !valid {
		return nil, false, nil
	}

	event_name, valid := log_params.Map().Get("ga_event_name")

	if !valid {
		return nil, false, nil
	}

	event := GAMPEvent{
		name:   event_name.AsString(),
		params: log_params.Map().AsRaw(),
	}

	return &event, true, nil
}

func logsToGALogs(logger *zap.Logger, ld plog.Logs) ([]*GAMPEvent, int) {
	n := ld.ResourceLogs().Len()
	if n == 0 {
		return []*GAMPEvent{}, 0
	}

	var dropped int
	var out []*GAMPEvent

	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)

		sls := rl.ScopeLogs()
		for j := 0; j < sls.Len(); j++ {
			sl := sls.At(j)
			logs := sl.LogRecords()
			for k := 0; k < logs.Len(); k++ {
				log := logs.At(k)
				event, valid, err := logToGAMPEvent(log)
				if !valid {
					dropped++
				}
				if err != nil {
					logger.Debug("Failed to convert to CloudWatch Log", zap.Error(err))
					dropped++
				} else {
					out = append(out, event)
				}
			}
		}
	}
	return out, dropped
}
