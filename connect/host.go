////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Contains functionality for describing and creating connections

package connect

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/crypto/signature/rsa"
	tlsCreds "gitlab.com/elixxir/crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"math"
	"sync"
	"time"
)

// Represents a reverse-authentication token
type Token []byte

// Information used to describe a connection to a host
type Host struct {
	// System-wide ID of the Host
	id string

	// address:Port being connected to
	address string

	// PEM-format TLS Certificate
	certificate []byte

	// Token shared with this Host establishing reverse authentication
	token Token

	// Configure the maximum number of connection attempts
	maxRetries int

	// GRPC connection object
	connection *grpc.ClientConn

	// TLS credentials object used to establish the connection
	credentials credentials.TransportCredentials

	// RSA Public Key corresponding to the TLS Certificate
	rsaPublicKey *rsa.PublicKey

	// If set, reverse authentication will be established with this Host
	enableAuth bool

	// Read/Write Mutex for thread safety
	mux sync.RWMutex
}

// Creates a new Host object
func NewHost(id, address string, cert []byte, disableTimeout,
	enableAuth bool) (host *Host, err error) {

	// Initialize the Host object
	host = &Host{
		id:          id,
		address:     address,
		certificate: cert,
		enableAuth:  enableAuth,
	}

	// Set the max number of retries for establishing a connection
	if disableTimeout {
		host.maxRetries = math.MaxInt32
	} else {
		host.maxRetries = 100
	}

	// Configure the host credentials
	err = host.setCredentials()
	return
}

// Checks if the given Host's connection is alive
func (h *Host) Connected() bool {
	h.mux.RLock()
	defer h.mux.RUnlock()

	return h.isAlive()
}

// CheckAndSend checks that the host has a connection and sends if it does.
// Operates under the host's read lock.
func (h *Host) send(f func(conn *grpc.ClientConn) (*any.Any,
	error)) (*any.Any, error) {

	h.mux.RLock()
	defer h.mux.RUnlock()

	if !h.isAlive() {
		return nil, errors.New("Could not send, connection is not alive")
	}

	a, err := f(h.connection)
	return a, err
}

// CheckAndStream checks that the host has a connection and streams if it does.
// Operates under the host's read lock.
func (h *Host) stream(f func(conn *grpc.ClientConn) (
	interface{}, error)) (interface{}, error) {

	h.mux.RLock()
	defer h.mux.RUnlock()

	if !h.isAlive() {
		return nil, errors.New("Could not stream, connection is not alive")
	}

	a, err := f(h.connection)
	return a, err
}

// Attempts to connect to the host if it does not have a valid connection
func (h *Host) connect() error {
	h.mux.Lock()
	defer h.mux.Unlock()

	//checks if the connection is active and skips reconnecting if it is
	if h.isAlive() {
		return nil
	}

	//connect to remote
	if err := h.connectHelper(); err != nil {
		return err
	}

	return nil
}

// authenticationRequired Checks if new authentication is required with
// the remote
func (h *Host) authenticationRequired() bool {
	h.mux.RLock()
	defer h.mux.RUnlock()

	return h.enableAuth && h.token==nil
}

// Checks if the given Host's connection is alive
func (h *Host) authenticate(handshake func(host *Host) error) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	return handshake(h)
}

// isAlive returns true if the connection is non-nil and alive
func (h *Host) isAlive() bool {
	if h.connection == nil {
		return false
	}
	state := h.connection.GetState()
	return state == connectivity.Idle || state == connectivity.Connecting ||
		state == connectivity.Ready
}

// Disconnect closes a the Host connection under the write lock
func (h *Host) Disconnect() {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.disconnect()
	h.token = nil
}

// disconnect closes a the Host connection while not under a write lock.
// undefined behavior if the caller has not taken the write lock
func (h *Host) disconnect() {
	// its possible to close a host which never sent so it never made a
	// connection. In that case, we should not close a connection which does not
	// exist
	if h.connection != nil {
		err := h.connection.Close()
		if err != nil {
			jww.ERROR.Printf("Unable to close connection to %s: %+v",
				h.address, errors.New(err.Error()))
		}
	}
}

// connect creates a connection while not under a write lock.
// undefined behavior if the caller has not taken the write lock
func (h *Host) connectHelper() (err error) {
	//attempts to disconnect to clean up an existing connection
	h.disconnect()
	// Configure TLS options
	var securityDial grpc.DialOption
	if h.credentials != nil {
		// Create the gRPC client with TLS
		securityDial = grpc.WithTransportCredentials(h.credentials)
	} else {
		// Create the gRPC client without TLS
		jww.WARN.Printf("Connecting to %v without TLS!", h.address)
		securityDial = grpc.WithInsecure()
	}

	// Attempt to establish a new connection
	for numRetries := 0; numRetries < h.maxRetries && !h.isAlive(); numRetries++ {

		jww.INFO.Printf("Connecting to address %+v. Attempt number %+v of %+v",
			h.address, numRetries, h.maxRetries)

		// If timeout is enabled, the max wait time becomes
		// ~14 seconds (with maxRetries=100)
		backoffTime := 2 * (numRetries/16 + 1)
		if backoffTime > 15 {
			backoffTime = 15
		}
		ctx, cancel := ConnectionContext(time.Duration(backoffTime))

		// Create the connection
		h.connection, err = grpc.DialContext(ctx, h.address, securityDial,
			grpc.WithBlock(), grpc.WithBackoffMaxDelay(time.Minute*5))
		if err != nil {
			jww.ERROR.Printf("Attempt number %+v to connect to %s failed: %+v\n",
				numRetries, h.address, errors.New(err.Error()))
		}
		cancel()
	}

	// Verify that the connection was established successfully
	if !h.isAlive() {
		return errors.New(fmt.Sprintf(
			"Last try to connect to %s failed. Giving up", h.address))
	}

	// Add the successful connection to the Manager
	jww.INFO.Printf("Successfully connected to %v", h.address)
	return
}

// Sets TransportCredentials and RSA PublicKey objects
// using a PEM-encoded TLS Certificate
func (h *Host) setCredentials() error {

	// If no TLS Certificate specified, print a warning and do nothing
	if h.certificate == nil || len(h.certificate) == 0 {
		jww.WARN.Printf("No TLS Certificate specified!")
		return nil
	}

	// Obtain the DNS name included with the certificate
	dnsName := ""
	cert, err := tlsCreds.LoadCertificate(string(h.certificate))
	if err != nil {
		s := fmt.Sprintf("Error forming transportCredentials: %+v", err)
		return errors.New(s)
	}
	if len(cert.DNSNames) > 0 {
		dnsName = cert.DNSNames[0]
	}

	// Create the TLS Credentials object
	h.credentials, err = tlsCreds.NewCredentialsFromPEM(string(h.certificate),
		dnsName)
	if err != nil {
		s := fmt.Sprintf("Error forming transportCredentials: %+v", err)
		return errors.New(s)
	}

	// Create the RSA Public Key object
	h.rsaPublicKey, err = tlsCreds.NewPublicKeyFromPEM(h.certificate)
	if err != nil {
		s := fmt.Sprintf("Error extracting PublicKey: %+v", err)
		return errors.New(s)
	}

	return nil
}

// Stringer interface for connection
func (h *Host) String() string {
	addr := h.address
	actualConnection := h.connection
	creds := h.credentials

	var state connectivity.State
	if actualConnection != nil {
		state = actualConnection.GetState()
	}

	serverName := "<nil>"
	protocolVersion := "<nil>"
	securityVersion := "<nil>"
	securityProtocol := "<nil>"
	if creds != nil {
		serverName = creds.Info().ServerName
		securityVersion = creds.Info().SecurityVersion
		protocolVersion = creds.Info().ProtocolVersion
		securityProtocol = creds.Info().SecurityProtocol
	}
	return fmt.Sprintf(
		"ID: %v\tAddr: %v\tCertificate: %v\tToken: %v\tEnableAuth: %v"+
			"\tMaxRetries: %v\tConnState: %v"+
			"\tTLS ServerName: %v\tTLS ProtocolVersion: %v\t"+
			"TLS SecurityVersion: %v\tTLS SecurityProtocol: %v\n",
		h.id, addr, h.certificate, h.token, h.enableAuth, h.maxRetries, state,
		serverName, protocolVersion, securityVersion, securityProtocol)
}
