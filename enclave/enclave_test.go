// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package enclave

import (
	"log"
	"net/http"
)

func ExampleCreateAttestationServerTLSConfig() {
	// Create a TLS config with a self-signed certificate and an embedded report.
	tlsConfig, err := CreateAttestationServerTLSConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTPS server.
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this is a test handler"))
	})
	server := http.Server{Addr: "0.0.0.0:8080", TLSConfig: tlsConfig}
	log.Fatal(server.ListenAndServeTLS("", ""))
}
