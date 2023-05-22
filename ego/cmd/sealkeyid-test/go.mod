module example.com/sealkeyid-test

go 1.19

replace github.com/edgelesssys/ego => ../../..

require github.com/edgelesssys/ego v1.3.0

require (
	golang.org/x/crypto v0.8.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)
