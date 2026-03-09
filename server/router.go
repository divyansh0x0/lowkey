package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	pb "github.com/ayush00git/lowkey/proto/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Router handles Redis Pub/Sub for message routing.
type Router struct {
	rdb *redis.Client
}

func NewRouter(addr string) *Router {
	return &Router{
		rdb: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// Publish sends a signaling message to a target peer via Redis.
func (r *Router) Publish(ctx context.Context, targetUUID string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	return r.rdb.Publish(ctx, fmt.Sprintf("signaling:%s", targetUUID), data).Err()
}

// Subscribe returns a channel of signaling messages for a specific peer.
func (r *Router) Subscribe(ctx context.Context, uuid string) *redis.PubSub {
	return r.rdb.Subscribe(ctx, fmt.Sprintf("signaling:%s", uuid))
}

// SignalPayload represents the data structure sent over Redis.
type SignalPayload struct {
	FromUUID string          `json:"from_uuid"`
	Type     string          `json:"type"` // "sdp" or "ice"
	Data     json.RawMessage `json:"data"`
}

func (r *Router) ForwardSDP(ctx context.Context, fromUUID string, sdp *pb.SdpExchange) error {
	data, _ := protojson.Marshal(sdp)
	payload := SignalPayload{
		FromUUID: fromUUID,
		Type:     "sdp",
		Data:     data,
	}
	return r.Publish(ctx, sdp.TargetUuid, payload)
}

func (r *Router) ForwardICE(ctx context.Context, fromUUID string, ice *pb.IceCandidate) error {
	data, _ := protojson.Marshal(ice)
	payload := SignalPayload{
		FromUUID: fromUUID,
		Type:     "ice",
		Data:     data,
	}
	return r.Publish(ctx, ice.TargetUuid, payload)
}
