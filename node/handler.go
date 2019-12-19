////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Contains callback interface for server functionality

package node

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/comms/connect"
	"gitlab.com/elixxir/comms/mixmessages"
	"google.golang.org/grpc/reflection"
	"runtime/debug"
)

// Server object used to implement endpoints and top-level comms functionality
type Comms struct {
	connect.ProtoComms
	handler Handler
}

// Starts a new server on the address:port specified by listeningAddr
// and a callback interface for server operations
// with given path to public and private key for TLS connection
func StartNode(localServer string, handler Handler,
	certPEMblock, keyPEMblock []byte) *Comms {
	pc, lis := connect.StartCommServer(localServer, certPEMblock, keyPEMblock)

	mixmessageServer := Comms{
		ProtoComms: pc,
		handler:    handler,
	}

	go func() {
		// Register GRPC services to the listening address
		mixmessages.RegisterNodeServer(mixmessageServer.LocalServer, &mixmessageServer)
		mixmessages.RegisterGenericServer(mixmessageServer.LocalServer, &mixmessageServer)

		// Register reflection service on gRPC server.
		reflection.Register(mixmessageServer.LocalServer)
		if err := mixmessageServer.LocalServer.Serve(lis); err != nil {
			err = errors.New(err.Error())
			jww.FATAL.Panicf("Failed to serve: %+v", err)
		}
		jww.INFO.Printf("Shutting down node server listener: %s", lis)
	}()

	return &mixmessageServer
}

type Handler interface {
	// Server interface for starting New Rounds
	CreateNewRound(message *mixmessages.RoundInfo) error
	// Server interface for sending a new batch
	PostNewBatch(message *mixmessages.Batch) error
	// Server interface for broadcasting when realtime is complete
	FinishRealtime(message *mixmessages.RoundInfo) error
	// GetRoundBufferInfo returns # of available precomputations
	GetRoundBufferInfo() (int, error)

	GetMeasure(message *mixmessages.RoundInfo) (*mixmessages.RoundMetrics, error)

	// Server Interface for all Internode Comms
	PostPhase(message *mixmessages.Batch)

	StreamPostPhase(server mixmessages.Node_StreamPostPhaseServer) error

	// Server interface for share broadcast
	PostRoundPublicKey(message *mixmessages.RoundPublicKey)

	// Server interface for RequestNonceMessage
	RequestNonce(salt []byte, RSAPubKey string, DHPubKey,
		RSASignedByRegistration, DHSignedByClientRSA []byte) ([]byte, []byte, error)

	// Server interface for ConfirmNonceMessage
	ConfirmRegistration(UserID []byte, Signature []byte) ([]byte, error)

	// PostPrecompResult interface to finalize both payloads' precomps
	PostPrecompResult(roundID uint64, slots []*mixmessages.Slot) error

	// GetCompletedBatch: gateway uses completed batch from the server
	GetCompletedBatch() (*mixmessages.Batch, error)

	PollNdf(ping *mixmessages.Ping) (*mixmessages.GatewayNdf, error)

	SendRoundTripPing(ping *mixmessages.RoundTripPing) error

	AskOnline(ping *mixmessages.Ping) error
}

type implementationFunctions struct {
	// Server Interface for starting New Rounds
	CreateNewRound func(message *mixmessages.RoundInfo) error
	// Server interface for sending a new batch
	PostNewBatch func(message *mixmessages.Batch) error
	// Server interface for finishing the realtime phase
	FinishRealtime func(message *mixmessages.RoundInfo) error
	// GetRoundBufferInfo returns # of available precomputations completed
	GetRoundBufferInfo func() (int, error)

	GetMeasure func(message *mixmessages.RoundInfo) (*mixmessages.RoundMetrics, error)

	// Server Interface for the Internode Messages
	PostPhase func(message *mixmessages.Batch)

	// Server interface for internode streaming messages
	StreamPostPhase func(message mixmessages.Node_StreamPostPhaseServer) error

	// Server interface for share broadcast
	PostRoundPublicKey func(message *mixmessages.RoundPublicKey)

	// Server interface for RequestNonceMessage
	RequestNonce func(salt []byte, RSAPubKey string, DHPubKey,
		RSASigFromReg, RSASigDH []byte) ([]byte, []byte, error)
	// Server interface for ConfirmNonceMessage
	ConfirmRegistration func(UserID, Signature []byte) ([]byte, error)

	// PostPrecompResult interface to finalize both payloads' precomputations
	PostPrecompResult func(roundID uint64,
		slots []*mixmessages.Slot) error

	GetCompletedBatch func() (*mixmessages.Batch, error)

	PollNdf func(ping *mixmessages.Ping) (*mixmessages.GatewayNdf, error)

	SendRoundTripPing func(ping *mixmessages.RoundTripPing) error

	AskOnline func(ping *mixmessages.Ping) error
}

// Implementation allows users of the client library to set the
// functions that implement the node functions
type Implementation struct {
	Functions implementationFunctions
}

// Below is the Implementation implementation, which calls the
// function matching the variable in the structure.

// NewImplementation returns a Implementation struct with all of the
// function pointers returning nothing and printing an error.
func NewImplementation() *Implementation {
	um := "UNIMPLEMENTED FUNCTION!"
	warn := func(msg string) {
		jww.WARN.Printf(msg)
		jww.WARN.Printf("%s", debug.Stack())
	}
	return &Implementation{
		Functions: implementationFunctions{
			CreateNewRound: func(m *mixmessages.RoundInfo) error {
				warn(um)
				return nil
			},
			PostPhase: func(m *mixmessages.Batch) {
				warn(um)
			},
			StreamPostPhase: func(message mixmessages.Node_StreamPostPhaseServer) error {
				warn(um)
				return nil
			},
			PostRoundPublicKey: func(message *mixmessages.RoundPublicKey) {
				warn(um)
			},
			PostNewBatch: func(message *mixmessages.Batch) error {
				warn(um)
				return nil
			},
			FinishRealtime: func(message *mixmessages.RoundInfo) error {
				warn(um)
				return nil
			},
			GetMeasure: func(message *mixmessages.RoundInfo) (*mixmessages.RoundMetrics, error) {
				warn(um)
				return nil, nil
			},
			GetRoundBufferInfo: func() (int, error) {
				warn(um)
				return 0, nil
			},

			RequestNonce: func(salt []byte, RSAPubKey string, DHPubKey,
				RSASig, RSASigDH []byte) ([]byte, []byte, error) {
				warn(um)
				return nil, nil, nil
			},
			ConfirmRegistration: func(UserID, Signature []byte) ([]byte, error) {
				warn(um)
				return nil, nil
			},
			PostPrecompResult: func(roundID uint64,
				slots []*mixmessages.Slot) error {
				warn(um)
				return nil
			},
			GetCompletedBatch: func() (batch *mixmessages.Batch, e error) {
				warn(um)
				return &mixmessages.Batch{}, nil
			},
			PollNdf: func(ping *mixmessages.Ping) (certs *mixmessages.GatewayNdf,
				e error) {
				warn(um)
				return &mixmessages.GatewayNdf{}, nil
			},
			SendRoundTripPing: func(ping *mixmessages.RoundTripPing) error {
				warn(um)
				return nil
			},
			AskOnline: func(ping *mixmessages.Ping) error {
				warn(um)
				return nil
			},
		},
	}
}

// Server Interface for starting New Rounds
func (s *Implementation) CreateNewRound(msg *mixmessages.RoundInfo) error {
	return s.Functions.CreateNewRound(msg)
}

func (s *Implementation) PostNewBatch(msg *mixmessages.Batch) error {
	return s.Functions.PostNewBatch(msg)
}

// Server Interface for the phase messages
func (s *Implementation) PostPhase(m *mixmessages.Batch) {
	s.Functions.PostPhase(m)
}

// Server Interface for streaming phase messages
func (s *Implementation) StreamPostPhase(m mixmessages.Node_StreamPostPhaseServer) error {
	return s.Functions.StreamPostPhase(m)
}

// Server Interface for the share message
func (s *Implementation) PostRoundPublicKey(message *mixmessages.
	RoundPublicKey) {
	s.Functions.PostRoundPublicKey(message)
}

// GetRoundBufferInfo returns # of completed precomputations
func (s *Implementation) GetRoundBufferInfo() (int, error) {
	return s.Functions.GetRoundBufferInfo()
}

// Server interface for RequestNonceMessage
func (s *Implementation) RequestNonce(salt []byte, RSAPubKey string, DHPubKey,
	RSASigFromReg, RSASigDH []byte) ([]byte, []byte, error) {
	return s.Functions.RequestNonce(salt, RSAPubKey, DHPubKey, RSASigFromReg, RSASigDH)
}

// Server interface for ConfirmNonceMessage
func (s *Implementation) ConfirmRegistration(UserID, Signature []byte) ([]byte, error) {
	return s.Functions.ConfirmRegistration(UserID, Signature)
}

// PostPrecompResult interface to finalize both payloads' precomputations
func (s *Implementation) PostPrecompResult(roundID uint64,
	slots []*mixmessages.Slot) error {
	return s.Functions.PostPrecompResult(roundID, slots)
}

func (s *Implementation) FinishRealtime(message *mixmessages.RoundInfo) error {
	return s.Functions.FinishRealtime(message)
}

func (s *Implementation) GetMeasure(message *mixmessages.RoundInfo) (*mixmessages.RoundMetrics, error) {
	return s.Functions.GetMeasure(message)
}

// Implementation of the interface using the function in the struct
func (s *Implementation) GetCompletedBatch() (*mixmessages.Batch, error) {
	return s.Functions.GetCompletedBatch()
}

func (s *Implementation) PollNdf(ping *mixmessages.Ping) (*mixmessages.
	GatewayNdf, error) {
	return s.Functions.PollNdf(ping)
}

func (s *Implementation) SendRoundTripPing(ping *mixmessages.RoundTripPing) error {
	return s.Functions.SendRoundTripPing(ping)
}

// AskOnline blocks until the server is online, or returns an error
func (s *Implementation) AskOnline(ping *mixmessages.Ping) error {
	return s.Functions.AskOnline(ping)
}
