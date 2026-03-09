package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ayush00git/lowkey/proto/v1"
)

var (
	targetUUID string
	sdpData    string
)

var rootCmd = &cobra.Command{
	Use:   "lowkey",
	Short: "Lowkey is a CLI for testing the signaling server",
	Long:  `A fast and flexible signaling CLI for WebRTC handshake simulation.`,
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Register and listen for incoming signaling messages",
	Run: func(cmd *cobra.Command, args []string) {
		id := uuid.New().String()
		fmt.Printf("My UUID: %s\n", id)

		conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewSignalingClient(conn)
		stream, err := client.Connect(context.Background())
		if err != nil {
			log.Fatalf("could not connect: %v", err)
		}

		// 1. Register
		err = stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Registration{
				Registration: &pb.Identity{
					Uuid: id,
				},
			},
		})
		if err != nil {
			log.Fatalf("registration failed: %v", err)
		}
		fmt.Println("Registered successfully. Waiting for signals...")

		// 2. Listen loop
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Fatalf("stream recv error: %v", err)
			}

			switch p := resp.Payload.(type) {
			case *pb.SignalResponse_Sdp:
				fmt.Printf("\n[SDP Received] Type: %s\nSDP: %s\nFrom Target: %s\n",
					p.Sdp.Type, p.Sdp.Sdp, p.Sdp.TargetUuid)
			case *pb.SignalResponse_Ice:
				fmt.Printf("\n[ICE Received] Candidate: %s\n", p.Ice.Candidate)
			case *pb.SignalResponse_Error:
				log.Printf("Server Error: %s", p.Error.Message)
			}
		}
	},
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a mock SDP Offer to a target peer",
	Run: func(cmd *cobra.Command, args []string) {
		if targetUUID == "" {
			log.Fatal("target UUID is required")
		}

		conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewSignalingClient(conn)
		stream, err := client.Connect(context.Background())
		if err != nil {
			log.Fatalf("could not connect: %v", err)
		}

		// 1. Identifying itself (sender)
		senderID := "cli-sender-" + uuid.New().String()[:8]
		err = stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Registration{
				Registration: &pb.Identity{
					Uuid: senderID,
				},
			},
		})
		if err != nil {
			log.Fatalf("registration failed: %v", err)
		}

		// 2. Send SDP Offer
		fmt.Printf("Sending SDP Offer to %s...\n", targetUUID)
		err = stream.Send(&pb.SignalRequest{
			Payload: &pb.SignalRequest_Sdp{
				Sdp: &pb.SdpExchange{
					Type:       pb.SdpExchange_TYPE_OFFER,
					Sdp:        sdpData,
					TargetUuid: targetUUID,
				},
			},
		})
		if err != nil {
			log.Fatalf("failed to send SDP: %v", err)
		}

		fmt.Println("SDP Offer sent successfully!")
		// Give the server a moment to receive the message before we close the stream.
		time.Sleep(500 * time.Millisecond)
	},
}

func init() {
	sendCmd.Flags().StringVarP(&targetUUID, "target", "t", "", "Target UUID to send signal to")
	sendCmd.Flags().StringVar(&sdpData, "sdp", "v=0\no=- 12345 12345 IN IP4 127.0.0.1\ns=-\nt=0 0\na=fingerprint:sha-256 ...", "Dummy SDP data")
	
	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(sendCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
