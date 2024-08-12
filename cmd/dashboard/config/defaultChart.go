package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gimlet-io/gimlet/pkg/dx"
)

type DefaultChart struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Chart       dx.Chart `json:"chart"`
}

type DefaultCharts []DefaultChart

func (c *DefaultCharts) Decode(value string) error {
	charts := []DefaultChart{}
	splittedCharts := strings.Split(value, ";")

	for _, chartsString := range splittedCharts {
		parsedChart, err := parseDefaultChartString(chartsString)
		if err != nil {
			return fmt.Errorf("invalid chart format: %s", err)
		}
		charts = append(charts, parsedChart)
	}
	*c = charts
	return nil
}

func (charts DefaultCharts) Find(chartName string) string {
	for _, c := range charts {
		if strings.Contains(c.Chart.Name, chartName) {
			return c.Chart.Version
		}
	}
	return ""
}

func (charts DefaultCharts) FindGitRepoHTTPSScheme(chart string) string {
	for _, c := range charts {
		if !strings.HasPrefix(c.Chart.Name, "git@") && !strings.Contains(c.Chart.Name, ".git") {
			continue
		}
		if strings.Contains(c.Chart.Name, chart) {
			return c.Chart.Name
		}
	}
	return ""
}

func parseDefaultChartString(chartsString string) (DefaultChart, error) {
	if chartsString == "" {
		return DefaultChart{}, nil
	}

	parsedValues, err := parse(chartsString)
	if err != nil {
		return DefaultChart{}, err
	}

	chart := dx.Chart{
		Name:       parsedValues.Get("name"),
		Repository: parsedValues.Get("repo"),
		Version:    parsedValues.Get("version"),
	}

	return DefaultChart{
		Title:       parsedValues.Get("title"),
		Description: parsedValues.Get("description"),
		Chart:       chart,
	}, nil
}

func parse(query string) (url.Values, error) {
	values := make(url.Values)
	err := populateValues(values, query)
	return values, err
}

func populateValues(values url.Values, query string) error {
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, ",")
		if strings.Contains(key, ";") {
			return fmt.Errorf("invalid semicolon separator in query")
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err := url.QueryUnescape(key)
		if err != nil {
			return err
		}
		value, err = url.QueryUnescape(value)
		if err != nil {
			return err
		}
		values[key] = append(values[key], value)
	}
	return nil
}
