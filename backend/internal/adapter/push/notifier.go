package push

import (
	"context"

	"xiaoheiplay/internal/domain"
	"xiaoheiplay/internal/usecase"
)

type OrderPushNotifier struct {
	orders usecase.OrderRepository
	push   *usecase.PushService
}

func NewOrderPushNotifier(orders usecase.OrderRepository, push *usecase.PushService) *OrderPushNotifier {
	return &OrderPushNotifier{
		orders: orders,
		push:   push,
	}
}

func (n *OrderPushNotifier) NotifyOrderEvent(ctx context.Context, ev domain.OrderEvent) error {
	if n.push == nil || n.orders == nil {
		return nil
	}
	if ev.Type != "order.pending_review" {
		return nil
	}
	order, err := n.orders.GetOrder(ctx, ev.OrderID)
	if err != nil {
		return err
	}
	return n.push.NotifyAdminsNewOrder(ctx, order)
}
