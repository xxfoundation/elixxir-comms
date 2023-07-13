////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// Contains remote sync gRPC endpoints

package server

import (
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/xx_network/comms/messages"
	"golang.org/x/net/context"
)

// Read data from the server
func (rc *Comms) Read(ctx context.Context, message *pb.RSReadRequest) (*pb.RSReadResponse, error) {
	return rc.handler.Read(message)
}

// Write data to the server
func (rc *Comms) Write(ctx context.Context, message *pb.RSWriteRequest) (*pb.RSWriteResponse, error) {
	return rc.handler.Write(message)
}

// GetLastModified returns the last time a resource was modified
func (rc *Comms) GetLastModified(ctx context.Context, message *pb.RSReadRequest) (*pb.RSTimestampResponse, error) {
	return rc.handler.GetLastModified(message)
}

// GetLastWrite returns the last time this remote sync server was modified
func (rc *Comms) GetLastWrite(ctx context.Context, message *messages.Ack) (*pb.RSTimestampResponse, error) {
	return rc.handler.GetLastWrite(message)
}

// ReadDir reads a directory from the server
func (rc *Comms) ReadDir(ctx context.Context, message *pb.RSReadRequest) (*pb.RSReadDirResponse, error) {
	return rc.handler.ReadDir(message)
}
