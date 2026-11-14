package logic

import (
	"errors"
	"testing"

	"github.com/chanify/chanify/model"
	"github.com/sideshow/apns2"
)

type deliveryPusher struct {
	responses []*apns2.Response
	err       error
	calls     int
}

func (p *deliveryPusher) Push(*apns2.Notification) (*apns2.Response, error) {
	i := p.calls
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	return p.responses[i], nil
}

func TestAPNSCountsOnlyAcceptedNotifications(t *testing.T) {
	previous := MockPusher
	t.Cleanup(func() { MockPusher = previous })
	for _, tc := range []struct {
		name      string
		responses []*apns2.Response
		err       error
		devices   []*model.Device
		timeline  bool
		want      int
		calls     int
	}{
		{"accepted", []*apns2.Response{{StatusCode: 200}}, nil, []*model.Device{{Type: 1}}, false, 1, 1},
		{"rejected", []*apns2.Response{{StatusCode: 400, Reason: "BadDeviceToken"}}, nil, []*model.Device{{Type: 1}}, false, 0, 1},
		{"missing response", []*apns2.Response{nil}, nil, []*model.Device{{Type: 1}}, false, 0, 1},
		{"network error", nil, errors.New("connection failed"), []*model.Device{{Type: 1}}, false, 0, 1},
		{"partial acceptance", []*apns2.Response{{StatusCode: 200}, {StatusCode: 410, Reason: "Unregistered"}}, nil, []*model.Device{{Type: 1}, {Type: 1}}, false, 1, 2},
		{"watch timeline skipped", nil, nil, []*model.Device{{Type: 2}}, true, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &deliveryPusher{responses: tc.responses, err: tc.err}
			MockPusher = p
			_, got := (&Logic{}).SendAPNS("synthetic-user", []byte("synthetic-message"), tc.devices, 10, "passive", tc.timeline)
			if got != tc.want || p.calls != tc.calls {
				t.Fatalf("accepted=%d calls=%d; want %d, %d", got, p.calls, tc.want, tc.calls)
			}
		})
	}
}
