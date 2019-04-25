////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Dummy implementation (so you can use for tests)
package node

import (
	pb "gitlab.com/elixxir/comms/mixmessages"
)

// Blank struct implementing ServerHandler interface for testing purposes (Passing to StartServer)
type TestInterface struct{}

func (m TestInterface) NewRound(roundId string) {}

func (m TestInterface) RoundtripPing(message *pb.TimePing) {}

func (m TestInterface) GetServerMetrics(message *pb.ServerMetrics) {}

func (m TestInterface) StartRound(messages *pb.Input) {}

func (m TestInterface) RunPhase(message *pb.Batch) {}
