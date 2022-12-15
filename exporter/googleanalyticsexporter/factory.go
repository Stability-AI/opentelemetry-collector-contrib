package googleanalyticsexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googleanalyticsexporter"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

const (
	typeStr = "googleanalytics"
	// The stability level of the exporter.
	stability = component.StabilityLevelDevelopment
)

// NewFactory creates a factory for GA exporter.
func NewFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		typeStr,
		createDefaultConfig,
		component.WithLogsExporter(createLogsExporter, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		ExporterSettings: config.NewExporterSettings(component.NewID(typeStr)),
	}
}

func createLogsExporter(_ context.Context, params component.ExporterCreateSettings, config component.Config) (component.LogsExporter, error) {
	expConfig, ok := config.(*Config)
	if !ok {
		return nil, errors.New("invalid configuration type; can't cast to googleanalyticslogsexporter.Config")
	}
	return newGALogsExporter(expConfig, params)
}
