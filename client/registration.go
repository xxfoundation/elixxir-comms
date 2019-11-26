////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Contains client -> registration server functionality

package client

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"gitlab.com/elixxir/comms/connect"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"google.golang.org/grpc"
)

// Client -> Registration Send Function
func (c *Comms) SendRegistrationMessage(host *connect.Host,
	message *pb.UserRegistration) (*pb.UserRegistrationConfirmation, error) {

	// Create the Send Function
	f := func(conn *grpc.ClientConn) (*any.Any, error) {
		// Set up the context
		ctx, cancel := connect.MessagingContext()
		defer cancel()

		// Send the message
		resultMsg, err := pb.NewRegistrationClient(conn).RegisterUser(ctx,
			message)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := host.Send(f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.UserRegistrationConfirmation{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// Client -> Registration Send Function
func (c *Comms) SendGetCurrentClientVersionMessage(
	host *connect.Host) (*pb.ClientVersion, error) {

	// Create the Send Function
	f := func(conn *grpc.ClientConn) (*any.Any, error) {
		// Set up the context
		ctx, cancel := connect.MessagingContext()
		defer cancel()

		// Send the message
		resultMsg, err := pb.NewRegistrationClient(
			conn).GetCurrentClientVersion(ctx, &pb.Ping{})
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := host.Send(f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.ClientVersion{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// Client -> Registration Send Function
func (c *Comms) SendGetUpdatedNDF(host *connect.Host,
	message *pb.NDFHash) (*pb.NDF, error) {

	// Create the Send Function
	f := func(conn *grpc.ClientConn) (*any.Any, error) {
		// Set up the context
		ctx, cancel := connect.MessagingContext()
		defer cancel()

		// Send the message
		resultMsg, err := pb.NewRegistrationClient(
			conn).PollNdf(ctx, message)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := host.Send(f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.NDF{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}
