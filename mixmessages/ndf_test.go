///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

package mixmessages

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"gitlab.com/xx_network/comms/messages"
	"gitlab.com/xx_network/comms/signature"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"reflect"
	"testing"
)

// Ensure message type conforms to genericSignable interface
// If this ever fails, check for modifications in the source library
//  as well as for this message type
var _ = signature.GenericSignable(&NDF{})

// ------------------------------------ SiGet/Set tests ------------------------------------------

// Happy path
func TestNDF_SetSignature(t *testing.T) {
	// Create ndf message
	testNdf := &NDF{}

	// Set the sig
	expectedSig := []byte("testSignature")
	expectedNonce := []byte("testNonce")
	testNdf.SetSignature(expectedSig, expectedNonce)

	expectedSignatureMsg := &messages.RSASignature{
		Nonce:     expectedNonce,
		Signature: expectedSig,
	}

	// Check that the ndf's signature is identical to the one set
	if !reflect.DeepEqual(expectedSignatureMsg, testNdf.Signature) {
		t.Errorf("Signature should match value it was set to! "+
			"Expected: %+v \n\t"+
			"Received: %+v", expectedSignatureMsg, testNdf.Signature)
	}
}

// Error path: attempt to set a nil signature or a nil nonce
func TestNDF_SetSignature_Eror(t *testing.T) {
	// Create ndf message
	testNdf := &NDF{}

	// Set the sig to nil (error case)
	nonNilValue := []byte("notNil")
	err := testNdf.SetSignature(nonNilValue, nil)
	if err == nil {
		t.Errorf("Expected error path: Should not be able to set signature as nil")
	}

	// Set the sig to nil (error case)
	err = testNdf.SetSignature(nil, nonNilValue)
	if err == nil {
		t.Errorf("Expected error path: Should not be able to set nonce as nil")
	}

}

// Happy path
func TestNDF_GetSignature(t *testing.T) {
	// Create roundErr and set signature (without using setSignature)
	expectedSig := []byte("expectedSig")
	expectedNonce := []byte("expectedNonce")
	expectedRsaSig := &messages.RSASignature{
		Signature: expectedSig,
		Nonce:     expectedNonce,
	}

	testNdf := &NDF{Signature: expectedRsaSig}

	// Fetch signature
	receivedSig := testNdf.GetSignature()

	// Compare fetched value to expected value
	if !reflect.DeepEqual(expectedRsaSig, receivedSig) {
		t.Errorf("Signature does not match one that was set!"+
			"Expected: %+v \n\t"+
			"Received: %+v", expectedRsaSig, receivedSig)
	}

}

// Error path (nil signature)
func TestNDF_GetSignature_NilCase(t *testing.T) {
	// Create ndf w/o signature object
	testNdf := &NDF{}

	// Attempt to get signature
	receivedSig := testNdf.GetSignature()

	// Received sig should be nil
	if receivedSig != nil {
		t.Errorf("Signature should default to nil if not set!")
	}

}

// -------------------- Digest tests -------------------------------

// Consistency test
func TestNDF_Digest_Consistency(t *testing.T) {
	// Generate a message
	expectedNdf := []byte("testNdf")
	testNdf := &NDF{Ndf: expectedNdf}

	// Hardcoded digest output. Any changes are a smoke test of changing of
	// crypto libraries
	expectedDigestEncoded := "mofeBfAPPiowXkQ/tVfzcccX2UZmqkzSqjxV6fM1avE="

	// Generate a digest
	sha := crypto.SHA256.New()
	testNonce := []byte("expectedNonce")
	digest := testNdf.Digest(testNonce, sha)

	// Encode outputted digest to base64 encoded string
	receivedDigestEncoded := base64.StdEncoding.EncodeToString(digest)

	// Check the consistency of generated digest and hard-coded digest
	if expectedDigestEncoded != receivedDigestEncoded {
		t.Errorf("Consistency test failed for testNDF."+
			"\n\tExpected: %v"+
			"\n\tRecieved: %v", expectedDigestEncoded, receivedDigestEncoded)
	}
}

// Test that digest output matches manual digest creation
func TestNDF_Digest(t *testing.T) {
	// Generate a message
	expectedNdf := []byte("testNdf")
	testNdf := &NDF{Ndf: expectedNdf}

	// Generate a digest
	sha := crypto.SHA256.New()
	testNonce := []byte("expectedNonce")
	receivedDigest := testNdf.Digest(testNonce, sha)

	// Manually generate the digest
	sha.Reset()
	sha.Write(expectedNdf)
	sha.Write(testNonce)
	expectedDigest := sha.Sum(nil)

	// Check that manual digest matches expected digest
	if !bytes.Equal(receivedDigest, expectedDigest) {
		t.Errorf("Digest did not output expected result."+
			"\n\tExpected: %v"+
			"\n\tRecieved: %v", expectedDigest, receivedDigest)
	}

}

// -------------------- Sign/Verify tests -------------------------------

// Happy path
func TestNdf_SignVerify(t *testing.T) {
	// Create ndf object
	ourNdf := []byte("testNdf")
	testNdf := &NDF{
		Ndf: ourNdf,
	}
	// Generate keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Errorf("Failed to generate key: %+v", err)
	}
	pubKey := privateKey.GetPublic()

	// Sign message
	err = signature.Sign(testNdf, privateKey)
	if err != nil {
		t.Errorf("Unable to sign message: %+v", err)
	}

	// Verify signature
	err = signature.Verify(testNdf, pubKey)
	if err != nil {
		t.Errorf("Expected happy path! Failed to verify: %+v", err)
	}
}

// Error path: Change internals of message between signing and verifying
func TestNdf_SignVerify_Error(t *testing.T) {
	// Create ndf object
	ourNdf := []byte("testNdf")
	testNdf := &NDF{
		Ndf: ourNdf,
	}

	// Generate keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Errorf("Failed to generate key: %+v", err)
	}
	pubKey := privateKey.GetPublic()

	// Sign message
	err = signature.Sign(testNdf, privateKey)
	if err != nil {
		t.Errorf("Unable to sign message: %+v", err)
	}

	// Reset ndf value so verify()'s signature won't match
	testNdf.Ndf = []byte("invalidChange")

	// Verify signature
	err = signature.Verify(testNdf, pubKey)
	// Verify signature
	if err == nil {
		t.Error("Expected error path: Should not have verified!")
	}

}
