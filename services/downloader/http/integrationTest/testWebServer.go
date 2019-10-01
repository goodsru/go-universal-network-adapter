package integrationTest

import (
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
)

// hasher uses package "crypto/sha256"
func hasher(s string) []byte {
	val := sha256.Sum256([]byte(s))
	return val[:]
}
func authHandler(handler http.HandlerFunc, userhash, passhash []byte, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare(hasher(user),
			userhash) != 1 || subtle.ConstantTimeCompare(hasher(pass), passhash) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}
func StartWebServer(serverType string) *http.Server {
	srv := &http.Server{}
	userhash := hasher("admin")
	passhash := hasher("$CrazyUnforgettablePassword?")
	realm := "Please enter username and password"
	r := http.NewServeMux()
	r.HandleFunc("/12345", indexHandler)
	r.HandleFunc("/basic/12345", authHandler(indexHandler, userhash, passhash, realm))
	switch serverType {
	case "HTTPS_CERT_VERIFY":
		go func() {
			caCert, err := ioutil.ReadFile("cert/rootCA1Cert.pem")
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			// Create the TLS Config with the CA pool and enable Client certificate validation
			tlsConfig := &tls.Config{
				ClientCAs:  caCertPool,
				ClientAuth: tls.RequireAndVerifyClientCert,
			}
			tlsConfig.BuildNameToCertificate()
			srv = &http.Server{Addr: ":9890", Handler: r}
			if err := srv.ListenAndServeTLS("cert/rootCA1Cert.pem", "cert/rootCA1Key.pem"); err != http.ErrServerClosed {
				panic(err)
			}
		}()
	case "HTTPS":
		go func() {
			srv = &http.Server{Addr: ":9889", Handler: r}
			if err := srv.ListenAndServeTLS("cert/rootCA1Cert.pem", "cert/rootCA1Key.pem"); err != http.ErrServerClosed {
				panic(err)
			}
		}()
	case "HTTP":
		go func() {
			srv = &http.Server{Addr: ":9888", Handler: r}
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				panic(err)
			}
		}()
	default:
		return nil
	}

	return srv
}
