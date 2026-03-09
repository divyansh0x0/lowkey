package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/ayush00git/lowkey/proto/v1"
)

type server struct {
	pb.UnimplementedSignalingServer
	router *Router
}

func newServer(redisAddr string) *server {
	return &server{
		router: NewRouter(redisAddr),
	}
}

func (s *server) Connect(stream pb.Signaling_ConnectServer) error {
	ctx := stream.Context()

	// 1. Initial Identity Registration
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	reg := req.GetRegistration()
	if reg == nil {
		return fmt.Errorf("initial message must be identity registration")
	}

	clientUUID := reg.Uuid
	log.Printf("Client connected: %s", clientUUID)

	// 2. Subscribe to Redis for this client
	pubsub := s.router.Subscribe(ctx, clientUUID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	// 3. Bidirectional logic
	errCh := make(chan error, 2)

	// Goroutine: Redis -> gRPC Client
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var payload SignalPayload
				if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
					log.Printf("Error unmarshaling redis payload: %v", err)
					continue
				}

				resp := &pb.SignalResponse{}
				switch payload.Type {
				case "sdp":
					var sdp pb.SdpExchange
					if err := protojson.Unmarshal(payload.Data, &sdp); err != nil {
						log.Printf("Error unmarshaling SDP: %v", err)
						continue
					}
					log.Printf("Forwarding SDP from Redis to client %s", clientUUID)
					resp.Payload = &pb.SignalResponse_Sdp{Sdp: &sdp}
				case "ice":
					var ice pb.IceCandidate
					if err := protojson.Unmarshal(payload.Data, &ice); err != nil {
						log.Printf("Error unmarshaling ICE: %v", err)
						continue
					}
					log.Printf("Forwarding ICE from Redis to client %s", clientUUID)
					resp.Payload = &pb.SignalResponse_Ice{Ice: &ice}
				}

				if err := stream.Send(resp); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()

	// Goroutine: gRPC Client -> Redis
	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				errCh <- nil
				return
			}
			if err != nil {
				errCh <- err
				return
			}

			switch p := req.Payload.(type) {
			case *pb.SignalRequest_Sdp:
				log.Printf("Client %s -> %s (SDP)", clientUUID, p.Sdp.TargetUuid)
				if err := s.router.ForwardSDP(ctx, clientUUID, p.Sdp); err != nil {
					log.Printf("Error forwarding SDP: %v", err)
				}
			case *pb.SignalRequest_Ice:
				log.Printf("Client %s -> %s (ICE)", clientUUID, p.Ice.TargetUuid)
				if err := s.router.ForwardICE(ctx, clientUUID, p.Ice); err != nil {
					log.Printf("Error forwarding ICE: %v", err)
				}
			}
		}
	}()

	return <-errCh
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSignalingServer(s, newServer("localhost:6379"))

	reflection.Register(s)

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
