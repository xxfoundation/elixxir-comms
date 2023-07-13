package client

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/messages"
)

// Read a resource from a RemoteSync server.
func (rc *Comms) Read(host *connect.Host, msg *pb.RSReadRequest) (*pb.RSReadResponse, error) {
	// Create the Send Function
	f := func(conn connect.Connection) (*any.Any, error) {
		// Set up the context
		ctx, cancel := host.GetMessagingContext()
		defer cancel()
		// Send the message
		resultMsg, err := pb.NewRemoteSyncClient(conn.GetGrpcConn()).
			Read(ctx, msg)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := rc.Send(host, f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.RSReadResponse{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// Write data to a path at a RemoteSync server
func (rc *Comms) Write(host *connect.Host, msg *pb.RSWriteRequest) (*pb.RSWriteResponse, error) {
	// Create the Send Function
	f := func(conn connect.Connection) (*any.Any, error) {
		// Set up the context
		ctx, cancel := host.GetMessagingContext()
		defer cancel()
		// Send the message
		resultMsg, err := pb.NewRemoteSyncClient(conn.GetGrpcConn()).
			Write(ctx, msg)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := rc.Send(host, f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.RSWriteResponse{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// GetLastModified returns the last time a path was modified.
func (rc *Comms) GetLastModified(host *connect.Host, msg *pb.RSReadRequest) (*pb.RSTimestampResponse, error) {
	// Create the Send Function
	f := func(conn connect.Connection) (*any.Any, error) {
		// Set up the context
		ctx, cancel := host.GetMessagingContext()
		defer cancel()
		// Send the message
		resultMsg, err := pb.NewRemoteSyncClient(conn.GetGrpcConn()).
			GetLastModified(ctx, msg)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := rc.Send(host, f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.RSTimestampResponse{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// GetLastWrite returns the last time a remote sync server was modified.
func (rc *Comms) GetLastWrite(host *connect.Host, msg *messages.Ack) (*pb.RSTimestampResponse, error) {
	// Create the Send Function
	f := func(conn connect.Connection) (*any.Any, error) {
		// Set up the context
		ctx, cancel := host.GetMessagingContext()
		defer cancel()
		// Send the message
		resultMsg, err := pb.NewRemoteSyncClient(conn.GetGrpcConn()).
			GetLastWrite(ctx, msg)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := rc.Send(host, f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.RSTimestampResponse{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}

// ReadDir returns all entries in a given path.
func (rc *Comms) ReadDir(host *connect.Host, msg *pb.RSReadRequest) (*pb.RSReadDirResponse, error) {
	// Create the Send Function
	f := func(conn connect.Connection) (*any.Any, error) {
		// Set up the context
		ctx, cancel := host.GetMessagingContext()
		defer cancel()
		// Send the message
		resultMsg, err := pb.NewRemoteSyncClient(conn.GetGrpcConn()).
			ReadDir(ctx, msg)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		return ptypes.MarshalAny(resultMsg)
	}

	// Execute the Send function
	resultMsg, err := rc.Send(host, f)
	if err != nil {
		return nil, err
	}

	// Marshall the result
	result := &pb.RSReadDirResponse{}
	return result, ptypes.UnmarshalAny(resultMsg, result)
}
