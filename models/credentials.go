// universal-network-adapter models
package models

import "crypto/tls"

// TLSMode represents the FTPS connection strategy. Servers cannot support
// both modes on the same port.
type TLSMode int

const (
	// TLSExplicit means the client first runs an explicit command ("AUTH TLS")
	// before switching to TLS.
	TLSExplicit TLSMode = 0

	// TLSImplicit means both sides already implicitly agree to use TLS, and the
	// client connects directly using TLS.
	TLSImplicit TLSMode = 1
)

// Credentials to authorize on remote server
type Credentials struct {
	// username
	User string
	// password
	Password string
	// sftp private key
	RsaPrivateKey string
	// sftp private key passphrase
	RsaPrivateKeyPassphrase string
	TLSConfig               *tls.Config
	TLSMode                 TLSMode
}
