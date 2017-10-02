package main

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
)

type Tag struct {
	Name  string
	Value string
}

type Metric struct {
	Tags  []Tag
	Value float64
}

func CreateMetrics(samples model.Vector) ([]Metric, error) {
	metrics := []Metric{}

	for _, sample := range samples {
		metric := Metric{}

		for name, value := range sample.Metric {
			tag := Tag{
				Name:  string(name),
				Value: string(value),
			}

			metric.Tags = append(metric.Tags, tag)
		}

		metric.Value = float64(sample.Value)

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func CreateGraphiteMetrics(samples model.Vector) (string, error) {
	metrics := ""

	for _, sample := range samples {
		name := sample.Metric["__name__"]

		value := sample.Value

		now := time.Now()
		timestamp := now.Unix()

		metric := fmt.Sprintf("%s %f %v\n", name, value, timestamp)

		metrics += metric
	}

	return metrics, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	promConfig := prometheus.Config{Address: "http://localhost:9090"}
	promClient, err := prometheus.New(promConfig)

	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

	promQueryClient := prometheus.NewQueryAPI(promClient)

	promResponse, err := promQueryClient.Query(ctx, "go_gc_duration_seconds", time.Now())

	if err != nil {
		fmt.Errorf("%v", err)
		return
	}

	if promResponse.Type() == model.ValVector {
		metrics, _ := CreateMetrics(promResponse.(model.Vector))
		for _, metric := range metrics {
			fmt.Printf("%+v\n", metric)
		}

		graphiteMetrics, _ := CreateGraphiteMetrics(promResponse.(model.Vector))
		fmt.Printf("%s\n", graphiteMetrics)
	}
}