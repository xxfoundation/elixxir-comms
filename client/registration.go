///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

// Contains client -> registration server functionality

package client

import (
	"crypto/sha256"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/ndf"
	"google.golang.org/grpc"
	"strings"
	"time"
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
	jww.DEBUG.Printf("Sending Registration message: %+v", message)
	resultMsg, err := c.Send(host, f)
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
			conn).GetCurrentClientVersion(ctx, &messages.Ping{})
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	jww.DEBUG.Printf("Sending Get Client Version message...")
	resultMsg, err := c.Send(host, f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.ClientVersion{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// RequestNdf is used to Request an ndf from permissioning
// Used by gateway, client, nodes and gateways
func (c *Comms) RequestNdf(host *connect.Host,
	message *pb.NDFHash) (*pb.NDF, error) {

	// Create the Send Function
	f := func(conn *grpc.ClientConn) (*any.Any, error) {
		// Set up the context
		ctx, cancel := connect.MessagingContext()
		defer cancel()

		authMsg, err := c.PackAuthenticatedMessage(message, host, false)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		// Send the message
		resultMsg, err := pb.NewRegistrationClient(
			conn).PollNdf(ctx, authMsg)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	jww.DEBUG.Printf("Sending Request Ndf message: %+v", message)
	resultMsg, err := c.Send(host, f)
	if err != nil {
		return nil, err
	}

	result := &pb.NDF{}
	return result, ptypes.UnmarshalAny(resultMsg, result)

}

// RetrieveNdf, attempts to connect to the permissioning server to retrieve the latest ndf for the notifications bot
func (c *Comms) RetrieveNdf(currentDef *ndf.NetworkDefinition) (*ndf.NetworkDefinition, error) {
	//Hash the notifications bot ndf for comparison with registration's ndf
	var ndfHash []byte
	// If the ndf passed not nil, serialize and hash it
	if currentDef != nil {
		//Hash the notifications bot ndf for comparison with registration's ndf
		hash := sha256.New()
		ndfBytes, err := currentDef.Marshal()
		if err != nil {
			return nil, err
		}
		hash.Write(ndfBytes)
		ndfHash = hash.Sum(nil)
	}
	//Put the hash in a message
	msg := &pb.NDFHash{Hash: ndfHash}

	regHost, ok := c.Manager.GetHost(&id.Permissioning)
	if !ok {
		return nil, errors.New("Failed to find permissioning host")
	}

	//Send the hash to registration
	response, err := c.RequestNdf(regHost, msg)

	// Keep going until we get a grpc error or we get an ndf
	for err != nil {
		// If there is an unexpected error
		if !strings.Contains(err.Error(), ndf.NO_NDF) {
			// If it is not an issue with no ndf, return the error up the stack
			errMsg := errors.Errorf("Failed to get ndf from permissioning: %v", err)
			return nil, errMsg
		}

		// If the error is that the permissioning server is not ready, ask again
		jww.WARN.Println("Failed to get an ndf, possibly not ready yet. Retying now...")
		time.Sleep(250 * time.Millisecond)
		response, err = c.RequestNdf(regHost, msg)

	}

	//If there was no error and the response is nil, client's ndf is up-to-date
	if response == nil || response.Ndf == nil {
		jww.DEBUG.Printf("Our NDF is up-to-date")
		return nil, nil
	}

	jww.INFO.Printf("Remote NDF: %s", string(response.Ndf))

	//Otherwise pull the ndf out of the response
	updatedNdf, _, err := ndf.DecodeNDF(string(response.Ndf))
	if err != nil {
		//If there was an error decoding ndf
		errMsg := errors.Errorf("Failed to decode response to ndf: %v", err)
		return nil, errMsg
	}
	return updatedNdf, nil
}
