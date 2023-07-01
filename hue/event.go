package hue

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tmaxmax/go-sse"
	"golang.org/x/exp/slog"
)

const (
	retrySleepDuration = 5 * time.Second
)

type Event struct {
	ID           string // The Hue event UUID.
	LastEventID  string // The SSE event ID.
	CreationTime time.Time
	Type         string
	Data         []Resource
}

type EventFilter func(Event) bool

type rawEvent struct {
	Event
	log *slog.Logger
}

func (r *rawEvent) UnmarshalJSON(data []byte) error {
	var rawData struct {
		ID           string            `json:"id"`
		CreationTime time.Time         `json:"creationtime"`
		Type         string            `json:"type"`
		Data         []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}

	r.ID = rawData.ID
	r.CreationTime = rawData.CreationTime
	r.Type = rawData.Type
	r.Data = nil

	for _, msg := range rawData.Data {
		var obj map[string]any
		if err := json.Unmarshal(msg, &obj); err != nil {
			return err
		}

		ty, ok := obj["type"]
		if !ok {
			r.log.Warn("Missing type field")
			continue
		}
		rTypeStr, ok := ty.(string)
		if !ok {
			r.log.Warn("Type field is not string")
			continue
		}
		rType := ResourceType(rTypeStr)

		var resource Resource
		switch rType {
		case RTypeLight:
			resource = &Light{}
		case RTypeScene:
			resource = &Scene{}
		default:
			r.log.Debug("Unknown resource type. Skipping", "type", rType)
			continue
		}

		if err := json.Unmarshal(msg, &resource); err != nil {
			return err
		}
		r.Data = append(r.Data, resource)
	}

	return nil
}

func (c *Client) EventListener(filter EventFilter, out chan<- Event) {
	// Listen for events forever. TODO: Do we ever want to manually cancel?
	for {
		err := c.listen(filter, out)
		c.log.Error("Error while listening for events. Retrying...",
			slog.Any("error", err),
			slog.Duration("retry_after", retrySleepDuration),
		)

		time.Sleep(retrySleepDuration)
	}
}

// Listen to events on http2 stream. Take a callback to use to filter events, and send
// matching events to a channel
// TODO: Does this need context?
func (c *Client) listen(filter EventFilter, out chan<- Event) error {
	req, _ := http.NewRequest(http.MethodGet, c.absURL("/eventstream/clip/v2"), nil)
	req.Header.Add(hueAppKeyHeader, c.AppKey)
	conn := c.sseClient.NewConnection(req)

	conn.SubscribeMessages(func(ev sse.Event) {
		if len(ev.Data) == 0 {
			return
		}

		c.log.Debug("Got event", "event", ev)

		var rawMsgs []json.RawMessage
		if err := json.Unmarshal(ev.Data, &rawMsgs); err != nil {
			c.log.Error("Error while unmarshalling message", "error", err)
			return
		}
		if len(rawMsgs) == 0 {
			return
		}

		for _, rawMsg := range rawMsgs {
			var raw rawEvent
			raw.log = c.log
			if err := json.Unmarshal(rawMsg, &raw); err != nil {
				c.log.Error("Error while unmarshalling message", slog.Any("error", err))
				continue
			}

			event := raw.Event
			event.LastEventID = ev.LastEventID
			if len(event.Data) == 0 || !filter(event) {
				continue
			}

			out <- event
		}
	})

	c.log.Info("Listening for bridge events")
	return conn.Connect()
	// TODO: Do we need to inject heartbeat events to ensure that connection is still alive?
	// See: https://github.com/home-assistant-libs/aiohue/blob/18ed0a4122bb9cf4f501267de26799a9e2c39dee/aiohue/v2/controllers/events.py#L284
}
