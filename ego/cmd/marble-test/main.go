// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"ego/test"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/edgelesssys/marblerun/coordinator/rpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	var t test.T
	defer t.Exit()
	require := require.New(&t)

	//
	// Run a mock Coordinator.
	//

	server := grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(generateCertificate())))
	rpc.RegisterMarbleServer(server, marbleServer{})

	listener, err := net.Listen("tcp", "localhost:")
	require.NoError(err)

	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()

	//
	// Run the test marble.
	//

	uuidDir, err := ioutil.TempDir("", "")
	require.NoError(err)
	defer os.RemoveAll(uuidDir)

	cmd := exec.Command("erthost", "ego-enclave:test-marble")
	cmd.Env = append(os.Environ(),
		"EDG_EGO_PREMAIN=1",
		"EDG_MARBLE_COORDINATOR_ADDR="+listener.Addr().String(),
		"EDG_MARBLE_TYPE=type",
		"EDG_MARBLE_DNS_NAMES=localhost",
		"EDG_MARBLE_UUID_FILE="+filepath.Join(uuidDir, "uuid"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL} // kill if parent dies
	require.NoError(cmd.Run())
}

type marbleServer struct{ rpc.UnimplementedMarbleServer }

func (marbleServer) Activate(context.Context, *rpc.ActivationReq) (*rpc.ActivationResp, error) {
	return &rpc.ActivationResp{Parameters: &rpc.Parameters{
		Argv: []string{"arg0", "arg1", "arg2"},
		Env:  map[string][]byte{"key1": []byte("val1"), "key2": []byte("val2")},
	}}, nil
}

func generateCertificate() *tls.Certificate {
	template := &x509.Certificate{
		SerialNumber: &big.Int{},
		Subject:      pkix.Name{CommonName: "localhost"},
		NotAfter:     time.Now().Add(time.Hour),
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert, _ := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	return &tls.Certificate{Certificate: [][]byte{cert}, PrivateKey: priv}
}
