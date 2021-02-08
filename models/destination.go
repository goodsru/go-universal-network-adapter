package models

import (
	"golang.org/x/crypto/ssh"
	goUrl "net/url"
	"time"
)

// Model of remote file/dir on remote server to be browsed/downloaded
type Destination struct {
	// remote file/dir url. Should start with protocol
	Url string
	// remote protocol. May be used to explicitly tell what protocol to use (i.e. "http", "ftp",
	// "etc").
	// Takes priority over protocol defined in Destination.Url
	Protocol string
	// credentials to access the remote file/dir
	Credentials *Credentials
	// connection timeout
	Timeout time.Duration
}

// Constructor for Destination
func NewDestination(url string, credentials *Credentials, timeout *time.Duration) *Destination {
	if timeout == nil {
		defaultTimeout := time.Duration(30) * time.Second
		timeout = &defaultTimeout
	}

	return &Destination{Url: url, Credentials: credentials, Timeout: *timeout}
}

// Processed Destination model for internal usage in universal network adapter
type ParsedDestination struct {
	// remote file/dir url
	Url string
	// remote protocol. May be used to explicitly tell what protocol to use (i.e. "http", "ftp",
	// "etc").
	// Takes priority over protocol defined in Destination.Url
	Protocol string
	// credentials to access the remote file/dir
	Credentials Credentials
	// Url parsed by using net/url parser
	ParsedUrl *goUrl.URL
	// connection timeout. Defaults to 30 seconds, if NewDestination is used
	Timeout time.Duration
}

// returns URL hostname
func (pd *ParsedDestination) GetHost() string {
	return pd.ParsedUrl.Host
}

// returns URL protocol
func (pd *ParsedDestination) GetScheme() string {
	return pd.ParsedUrl.Scheme
}

// returns URL path
func (pd *ParsedDestination) GetPath() string {
	return pd.ParsedUrl.Path
}

// returns user from parsed Credentials
func (pd *ParsedDestination) GetUser() string {
	return pd.Credentials.User
}

// returns password from parsed Credentials
func (pd *ParsedDestination) GetPassword() string {
	return pd.Credentials.Password
}

// returns RSA private key for sftp connect from parsed Credentials
func (pd *ParsedDestination) GetRsaPrivateKey() (ssh.Signer, error) {
	if pd.Credentials.RsaPrivateKeyPassphrase == "" {
		return ssh.ParsePrivateKey([]byte(pd.Credentials.RsaPrivateKey))
	}
	return ssh.ParsePrivateKeyWithPassphrase([]byte(pd.Credentials.RsaPrivateKey), []byte(pd.Credentials.RsaPrivateKeyPassphrase))
}

// parses Destination to ParsedDestination
func ParseDestination(destination *Destination) (*ParsedDestination, error) {
	parsedUrl, err := goUrl.Parse(destination.Url)
	if err != nil {
		return nil, err
	}

	credentials := destination.Credentials
	if credentials == nil {
		credentials = &Credentials{}
		credentials.User = parsedUrl.User.Username()
		if pass, ok := parsedUrl.User.Password(); ok {
			credentials.Password = pass
		}
	}

	parsedUrl.User = nil

	return &ParsedDestination{Url: parsedUrl.String(), Protocol: destination.Protocol, Credentials: *credentials, ParsedUrl: parsedUrl, Timeout: destination.Timeout}, nil
}
