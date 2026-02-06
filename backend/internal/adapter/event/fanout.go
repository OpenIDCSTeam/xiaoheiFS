package event

import (
	"context"

	"xiaoheiplay/internal/domain"
	"xiaoheiplay/internal/usecase"
)

type EventSink interface {
	NotifyOrderEvent(ctx context.Context, ev domain.OrderEvent) error
}

type FanoutPublisher struct {
	primary usecase.EventPublisher
	sinks   []EventSink
}

func NewFanoutPublisher(primary usecase.EventPublisher, sinks ...EventSink) *FanoutPublisher {
	return &FanoutPublisher{primary: primary, sinks: sinks}
}

func (p *FanoutPublisher) Publish(ctx context.Context, orderID int64, eventType string, payload any) (domain.OrderEvent, error) {
	ev, err := p.primary.Publish(ctx, orderID, eventType, payload)
	if err != nil {
		return ev, err
	}
	for _, sink := range p.sinks {
		if sink == nil {
			continue
		}
		_ = sink.NotifyOrderEvent(ctx, ev)
	}
	return ev, nil
}

var _ usecase.EventPublisher = (*FanoutPublisher)(nil)
