////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Contains server gRPC endpoints

package node

// TODO: A lot of message types from gRPC are passed through, and a number of
//       errors that can occur are not accounted for.

import (
	pb "gitlab.com/elixxir/comms/mixmessages"
	"golang.org/x/net/context"
)

// Handle a Broadcasted Ask Online event
func (s *NodeComms) AskOnline(ctx context.Context, msg *pb.Ping) (
	*pb.Ack, error) {
	return &pb.Ack{}, nil
}

// Handle a Roundtrip ping event
func (s *NodeComms) RoundtripPing(ctx context.Context, msg *pb.TimePing) (
	*pb.Ack, error) {
	serverHandler.RoundtripPing(msg)
	return &pb.Ack{}, nil
}

// Handle a broadcasted ServerMetric event
func (s *NodeComms) GetServerMetrics(ctx context.Context, msg *pb.ServerMetrics) (
	*pb.Ack, error) {
	serverHandler.GetServerMetrics(msg)
	return &pb.Ack{}, nil
}

// Handle a NewRound event
func (s *NodeComms) CreateNewRound(ctx context.Context,
	msg *pb.RoundInfo) (*pb.Ack, error) {
	// Call the server handler to start a new round
	serverHandler.CreateNewRound(msg)
	return &pb.Ack{}, nil
}

// PostNewBatch polls the first node and sends a batch when it is ready
func (s *NodeComms) PostNewBatch(ctx context.Context, msg *pb.Batch) (*pb.Ack, error) {
	// Call the server handler to post a new batch
	serverHandler.PostNewBatch(msg)

	return &pb.Ack{}, nil
}

// Handle a Phase event
func (s *NodeComms) PostPhase(ctx context.Context, msg *pb.Batch) (*pb.Ack,
	error) {
	// Call the server handler with the msg
	serverHandler.PostPhase(msg)
	return &pb.Ack{}, nil
}

// Handle a PostRoundPublicKey message
func (s *NodeComms) PostRoundPublicKey(ctx context.Context,
	msg *pb.RoundPublicKey) (*pb.Ack, error) {
	// Call the server handler that receives the key share
	serverHandler.PostRoundPublicKey(msg)
	return &pb.Ack{}, nil
}

// Handle a StartRealtime event
func (s *NodeComms) StartRealtime(ctx context.Context, msg *pb.Batch) (
	*pb.Ack, error) {
	serverHandler.PostNewBatch(msg)
	return &pb.Ack{}, nil
}

// GetBufferInfo returns buffer size (number of completed precomputations)
func (s *NodeComms) GetRoundBufferInfo(ctx context.Context,
	msg *pb.RoundBufferInfo) (
	*pb.RoundBufferInfo, error) {
	bufSize, err := serverHandler.GetRoundBufferInfo()
	if bufSize < 0 {
		bufSize = 0
	}
	size := uint32(bufSize)
	return &pb.RoundBufferInfo{RoundBufferSize: size}, err
}

// Handles Registration Nonce Communication
func (s *NodeComms) RequestNonce(ctx context.Context,
	msg *pb.NonceRequest) (*pb.Nonce, error) {
	pk := msg.GetClient()
	sig := msg.GetClientSignedByServer()

	// Obtain the nonce by passing to server
	nonce, err := serverHandler.RequestNonce(msg.GetSalt(),
		pk.GetY(), pk.GetP(), pk.GetQ(),
		pk.GetG(), sig.GetHash(), sig.GetR(), sig.GetS())

	// Obtain the error message, if any
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	// Return the NonceMessage
	return &pb.Nonce{
		Nonce: nonce,
		Error: errMsg,
	}, err
}

// Handles Registration Nonce Confirmation
func (s *NodeComms) ConfirmRegistration(ctx context.Context,
	msg *pb.DSASignature) (*pb.RegistrationConfirmation, error) {

	// Obtain signed client public key by passing to server
	hash, R, S, Y, P, Q, G, err := serverHandler.ConfirmRegistration(
		msg.GetHash(),
		msg.GetR(), msg.GetS())

	// Obtain the error message, if any
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	// Return the RegistrationConfirmation
	return &pb.RegistrationConfirmation{
		ClientSignedByServer: &pb.DSASignature{
			Hash: hash,
			R:    R,
			S:    S,
		},
		Server: &pb.DSAPublicKey{
			Y: Y,
			P: P,
			Q: Q,
			G: G,
		},
		Error: errMsg,
	}, err
}

// PostPrecompResult sends final Message and AD precomputations.
func (s *NodeComms) PostPrecompResult(ctx context.Context,
	msg *pb.Batch) (*pb.Ack, error) {
	// Call the server handler to start a new round
	err := serverHandler.PostPrecompResult(msg.GetRound().GetID(),
		msg.GetSlots())
	return &pb.Ack{}, err
}

// FinishRealtime broadcasts to all nodes when the realtime is completed
func (s *NodeComms) FinishRealtime(ctx context.Context, msg *pb.Ping) (*pb.Ack, error) {
	// Call the server handler to finish realtime
	err := serverHandler.FinishRealtime()

	return &pb.Ack{}, err
}
