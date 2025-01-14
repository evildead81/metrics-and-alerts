package instance

import (
	"context"
	"fmt"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	pb "github.com/evildead81/metrics-and-alerts/internal/proto/metrics/proto"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

type server struct {
	pb.UnimplementedMetricsServiceServer
	storage storages.Storage
	key     string
}

func (s *server) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]contracts.Metrics, len(req.Metrics))
	for i, m := range req.Metrics {
		metrics[i] = contracts.Metrics{
			ID:    m.Id,
			MType: m.Type,
			Delta: &m.Delta,
			Value: &m.Value,
		}
	}

	if err := s.storage.UpdateMetrics(metrics); err != nil {
		return &pb.UpdateMetricsResponse{Success: false}, err
	}
	return &pb.UpdateMetricsResponse{Success: true}, nil
}

func (s *server) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	metric := contracts.Metrics{
		ID:    req.Metric.Id,
		MType: req.Metric.Type,
		Delta: &req.Metric.Delta,
		Value: &req.Metric.Value,
	}

	newMetric := contracts.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	switch {
	case metric.MType == consts.Gauge:
		err := s.storage.UpdateGauge(metric.ID, *metric.Value)
		if err != nil {
			return nil, err
		}
		newMetric.Value = metric.Value
	case metric.MType == consts.Counter:
		err := s.storage.UpdateCounter(metric.ID, *metric.Delta)
		if err != nil {
			return nil, err
		}
		updatedCounterValue, err := s.storage.GetCountValueByName(metric.ID)
		if err != nil {
			return nil, err
		}
		newMetric.Delta = &updatedCounterValue
	default:
		return nil, fmt.Errorf("incorrect type")
	}

	return &pb.UpdateMetricResponse{Metric: &pb.Metrics{
		Id:    newMetric.ID,
		Type:  newMetric.MType,
		Delta: *newMetric.Delta,
		Value: *newMetric.Value,
	}}, nil
}
