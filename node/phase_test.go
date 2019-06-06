////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

package node

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"gitlab.com/elixxir/comms/connect"
	"gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/comms/testkeys"
	"google.golang.org/grpc/metadata"
	"io"
	"reflect"
	"testing"
)

func TestPhase_StreamPostPhase(t *testing.T) {

	// Init server receiver
	servReceiverAddress := getNextServerAddress()
	receiverImpl := NewImplementation()
	receiverImpl.Functions.StreamPostPhase = func(server mixmessages.Node_StreamPostPhaseServer) error {
		return mockStreamPostPhase(server)
	}
	serverStreamReceiver := StartNode(servReceiverAddress, receiverImpl,
		testkeys.GetNodeCertPath(), testkeys.GetNodeKeyPath())

	// Init server sender
	servSenderAddress := getNextServerAddress()
	serverStreamSender := StartNode(servSenderAddress, NewImplementation(),
		testkeys.GetNodeCertPath(), testkeys.GetNodeKeyPath())

	// Get credentials and connect to node
	creds := connect.NewCredentialsFromFile(testkeys.GetNodeCertPath(),
		"*.cmix.rip")

	senderToReceiverID := MockID("sender2receiver")
	receiverToSenderID := MockID("receiver2tosender")
	// It might make more sense to call the RPC on the connection object
	// that's returned from this
	serverStreamSender.ConnectToNode(senderToReceiverID, servReceiverAddress, creds)
	serverStreamSender.ConnectToNode(receiverToSenderID, servSenderAddress, creds)

	// Reset TLS-related global variables
	defer serverStreamReceiver.Shutdown()
	defer serverStreamSender.Shutdown()

	// Create streaming context so you can close stream later
	ctx, cancel := connect.StreamingContext()

	// Create round info to be used in batch info
	roundId := uint64(10)
	roundInfo := mixmessages.RoundInfo{
		ID: roundId,
	}

	// Create header which contains the batch info
	forPhase := int32(3)
	batchInfo := mixmessages.BatchInfo{
		Round:    &roundInfo,
		ForPhase: forPhase,
	}

	// Create a new context with some metadata
	// using the batch info header
	//headerBuffer, err := proto.Marshal(&batchInfo)
	ctx = metadata.AppendToOutgoingContext(ctx, "batchinfo", batchInfo.String())

	// Get stream client for post phase
	streamClient, err := serverStreamSender.GetPostPhaseStream(senderToReceiverID, ctx)

	if err != nil {
		t.Errorf("Unable to get streamling clinet %v", err)
	}

	// Generate indexed slots
	numSlots := 3
	slots := createSlots(numSlots)

	// The server will send the slot messages and wait for ack
	// and close on receiving it.
	go func(slots []mixmessages.Slot) {
		for i, slot := range slots {
			err := streamClient.Send(&slot)
			if err != nil {
				t.Errorf("Unable to send slot %v %v", i, err)
			}
		}
		ack, err := streamClient.CloseAndRecv()

		if err != nil {
			t.Errorf("Failed to close and receive stream")
		}

		if ack != nil && ack.Error != "" {
			t.Errorf("Remote Server Error in ack: %s", ack.Error)
		}

	}(slots)

	// Block until stream client context is cleaned up
	<-streamClient.Context().Done()

	// Compared received slots to expected values
	for i, slot := range slots {
		if !reflect.DeepEqual(*receivedBatch.Slots[i], slot) {
			t.Errorf("Received slot %v doesn't match expected value: %v, got %v", i, slot, *receivedBatch.Slots[i])
		}
	}

	// Compare for phase
	if forPhase != receivedBatch.ForPhase {
		t.Errorf("ForPhase received %v doesn't match expected %v", receivedBatch.ForPhase, forPhase)
	}

	// Compare round info
	if !reflect.DeepEqual(roundInfo, *receivedBatch.Round) {
		t.Errorf("Round info received %v doesn't match expected %v", *receivedBatch.Round, roundInfo)
	}

	// Clean up sender context
	cancel()
}

var receivedBatch mixmessages.Batch

func mockStreamPostPhase(stream mixmessages.Node_StreamPostPhaseServer) error {

	// Unmarshal header into batch info
	batchInfo := mixmessages.BatchInfo{}
	md, ok := metadata.FromIncomingContext(stream.Context())

	if !ok {
		return errors.New("unable to retrieve meta data / header %v")
	}

	err := proto.UnmarshalText(md.Get("batchinfo")[0], &batchInfo)
	if err != nil {
		return err
	}

	// Create batch using batch info header
	receivedBatch = mixmessages.Batch{
		Round:    batchInfo.Round,
		ForPhase: batchInfo.ForPhase,
		Slots:    make([]*mixmessages.Slot, 3),
	}

	// Send a chunk per received slot until EOF
	index := uint32(0)
	for {
		slot, err := stream.Recv()
		// If we are at end of receiving
		// send ack and finish
		if err == io.EOF {
			ack := mixmessages.Ack{
				Error: "",
			}
			err = stream.SendAndClose(&ack)
			return err
		}

		// If we have another error, return err
		if err != nil {
			return err
		}

		// Store slot received into received batch
		receivedBatch.Slots[index] = slot
		index++
	}
}

func createSlots(numSlots int) []mixmessages.Slot {

	slots := make([]mixmessages.Slot, numSlots)

	for i := 0; i < numSlots; i++ {
		slots[i] = mixmessages.Slot{
			MessagePayload: []byte{0x01},
			SenderID:       []byte{0x02},
		}
	}

	return slots
}
