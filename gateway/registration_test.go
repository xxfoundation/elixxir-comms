////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package gateway

import (
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/comms/node"
	"gitlab.com/xx_network/comms/connect"
	"gitlab.com/xx_network/comms/gossip"
	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/primitives/id"
	"testing"
)

// Smoke test SendRequestClientKeyMessage
func TestSendRequestNonceMessage(t *testing.T) {
	GatewayAddress := getNextGatewayAddress()
	ServerAddress := getNextServerAddress()
	testID := id.NewIdFromString("test", id.Generic, t)
	gateway := StartGateway(testID, GatewayAddress, NewImplementation(), nil,
		nil, gossip.DefaultManagerFlags())
	server := node.StartNode(testID, ServerAddress, 0, node.NewImplementation(),
		nil, nil)
	defer gateway.Shutdown()
	defer server.Shutdown()
	manager := connect.NewManagerTesting(t)

	params := connect.GetDefaultHostParams()
	params.AuthEnabled = false
	host, err := manager.AddHost(testID, ServerAddress, nil, params)
	if err != nil {
		t.Errorf("Unable to call NewHost: %+v", err)
	}

	RSASignature := &messages.RSASignature{
		Signature: []byte("test"),
	}

	_, err = gateway.SendRequestClientKeyMessage(host,
		&pb.SignedClientKeyRequest{ClientKeyRequestSignature: RSASignature})
	if err != nil {
		t.Errorf("SendRequestClientKeyMessage: Error received: %s", err)
	}
}

func TestPoll(t *testing.T) {
	GatewayAddress := getNextGatewayAddress()
	ServerAddress := getNextServerAddress()

	testID := id.NewIdFromString("test", id.Generic, t)
	gateway := StartGateway(testID, GatewayAddress, NewImplementation(), nil,
		nil, gossip.DefaultManagerFlags())
	server := node.StartNode(testID, ServerAddress, 0, node.NewImplementation(),
		nil, nil)
	defer gateway.Shutdown()
	defer server.Shutdown()
	manager := connect.NewManagerTesting(t)

	params := connect.GetDefaultHostParams()
	params.AuthEnabled = false
	host, err := manager.AddHost(testID, ServerAddress, nil, params)
	if err != nil {
		t.Errorf("Unable to call NewHost: %+v", err)
	}

	_, err = gateway.SendPoll(host, &pb.ServerPoll{})
	if err != nil {
		t.Errorf("TestDemandNdf: Error received: %s", err)
	}
}
