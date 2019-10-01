package sftp

//see https://github.com/pkg/sftp/blob/master/server_integration_test.go ....

import (
	"bytes"
	"encoding/hex"
	"github.com/pkg/sftp"
	"io"
	"io/ioutil"
	"log"

	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

// starts an ssh server to test. returns: host string and port
func startTestSftpServer(port int, login string, password string, pubKeyRsa string, isDebug bool) (net.Listener, string, int) {
	debugStream := ioutil.Discard
	if isDebug {
		debugStream = os.Stderr
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", port))
	if err != nil {
		panic(err)
	}

	host, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		panic(err)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	config := &ssh.ServerConfig{}

	if login != "" {
		config.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			fmt.Fprintf(debugStream, "Login: %s\n", c.User())
			if c.User() == login && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		}
	}

	if pubKeyRsa != "" {
		config.PublicKeyCallback = func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {

			pubKeyParsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubKeyRsa))
			if err != nil {
				log.Fatal(err)
			}

			expected := pubKeyParsed.Marshal()
			fact := pubKey.Marshal()

			if bytes.Equal(expected, fact) {
				return nil, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		}
	}

	private, err := ssh.ParsePrivateKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAkTMV7QOcMKwJj8g9CLr5p09TP82lu8F/nfs3csK6Ft2vfev0
GJKAZAaIk3BjdQzOKeT7bpXdbofjF6NGEOgqG7IneKIuDNlt6Rr/pnQtwtFlq8iH
Qzj2kjkMN7AoHvcQ2O/FJD5I494E2Emn/CdreX1QSUIzZCfox4IQH/Iygwd28+zT
+IZgVLLyqB9xogYqRtY/aCRPFlXsVll5b9yYI5Po2Ztm3GKHR725PFC+gNZkqaL6
IO4qWJzZ+m6WWyFn+2cLs2AAz/FWDEeiGyrSbbkDxRG65OibuIexE26rl6FFxUtO
iHxe11WKHxvIMfTIXbwn/9E0ijbFkjQ9np2JMQIDAQABAoIBACyqSlReh+1w/n/8
WMoWujV9cV5vJExdeaRfYu8qd5QRHareqnfkmRor6FdyeyXGiqTYi3+5GiSmKHVd
BHCu+kUKyJmTShcpy2WpqHrLwkfrohr11KYZH8BbTCbVSYI8/jG9kCdWAZYW9eaS
wdoPRkBZDBl9A9ILdM/qcothNaiYaj4ex1BE8H8zOfE06kuS3wWuWD/rOKTKIj1L
2UlTAzY+VKATctcTqjwmbCBuTM9/74QlcuekyYuL5kb4Tp5vfO/hy71XgjCVpAjC
/OvKIb1DzJ4d1A6WaGiQ09bMeoE7j0+sUpwOZpXV1PBa328iZeFWCymSJZgFwae8
B8VAwwECgYEA649O0vuCus0l1NfYqUv4DBzXodceQE/z+jEDtMtc0g2YEKkTkvbJ
BFBXh6SUWGu6maRd56VFCUu7S6l7zmWpxVY7WixMh+tLTdzem1aPoy770PnYVgIq
oE1n+wdbT9an2Xm8RyQc4rN0DRki2JjlYYKQEnc2uN/Yo3NzgoqgEZ8CgYEAncyH
LebqcXbX+MeN82coZV6mdsg+OnUT1mt0XSk4GN0TRlSX12ucvue/FD+hAVX2a+IO
vCu3Ig/cDAt7Ck+zbWvkO0p9IQTamHWel6WiU7j61oc7Me80wBrAq8qmPLJ4zn0w
Ug3OyoMDFBtO4aeqXD5fiCEEwQWm8zf51EFkky8CgYEAo7hxArcIf6kCOFLlJZXF
izWosbWAMxbe7e+PMeN+WghUvo+lKSoZQMROcdpzuJj3kr3o/o2h0/os7UOY5zVF
sABlNCFB994T8sQmfDTHlJWdM/vS3sSWt9/U+2Z7kpwRAVhZAeEZqn8rk8b6ryxl
kUZbuFnfUNNUERa3G+4ZnuUCgYEAi+MuyyB0IVYLq72sN2TxyLuZsp9dmxEDHwhv
Rc8urcv+NFD1ssDxWcOz/s1RfA+qvoTOLz5JwOZyWjMrRj7Vf2EwGOe1+bmF17Yd
e64YM0Q/CkMj1OaLyulseF7T8+b7dYJBfdxDv/9YkVCMIzsxqUiaA+HRhxPtppuu
292EvX0CgYBtBjuCeEbNSMN8b9OHYHBexTqBkx8nDHTNW61R7cBo6gH30jAKmLSm
5uA1MemHJXEFtYm42g9fVai6ydWLSb44WISID1G5oF+/Nh/1j93aGSBBzFXq9ccH
MMpP/Pinco5UsoxgxQYBFD2HgGWt3FEuTG5p1ACmsDvl7MnXOJgmRQ==
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		log.Fatal("Failed to parse private key", err)
	}

	config.AddHostKey(private)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Fprintf(debugStream, "ssh server socket closed: %v\n", err)
				break
			}

			go func() {
				defer conn.Close()
				sshSvr, err := sshServerFromConn(conn, config, debugStream)
				if err != nil {
					fmt.Fprintf(debugStream, "ssh server error: %v\n", err)
					return
				}
				err = sshSvr.Wait()
				fmt.Fprintf(debugStream, "ssh server finished, err: %v\n", err)
			}()
		}
	}()

	return listener, host, port
}

type sshServer struct {
	conn     net.Conn
	config   *ssh.ServerConfig
	sshConn  *ssh.ServerConn
	newChans <-chan ssh.NewChannel
	newReqs  <-chan *ssh.Request
}

func sshServerFromConn(conn net.Conn, config *ssh.ServerConfig, debugStream io.Writer) (*sshServer, error) {
	// From a standard TCP connection to an encrypted SSH connection
	sshConn, newChans, newReqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return nil, err
	}

	svr := &sshServer{conn, config, sshConn, newChans, newReqs}
	svr.listenChannels(debugStream)
	return svr, nil
}

func (svr *sshServer) Wait() error {
	return svr.sshConn.Wait()
}

func (svr *sshServer) Close() error {
	return svr.sshConn.Close()
}

func (svr *sshServer) listenChannels(debugStream io.Writer) {
	go func() {
		for chanReq := range svr.newChans {
			go svr.handleChanReq(chanReq, debugStream)
		}
	}()
	go func() {
		for req := range svr.newReqs {
			go svr.handleReq(req, debugStream)
		}
	}()
}

func (svr *sshServer) handleReq(req *ssh.Request, debugStream io.Writer) {
	switch req.Type {
	default:
		rejectRequest(req, debugStream)
	}
}

type sshChannelServer struct {
	svr     *sshServer
	chanReq ssh.NewChannel
	ch      ssh.Channel
	newReqs <-chan *ssh.Request
}

type sshSessionChannelServer struct {
	*sshChannelServer
	env []string
}

func (svr *sshServer) handleChanReq(chanReq ssh.NewChannel, debugStream io.Writer) {
	fmt.Fprintf(debugStream, "channel request: %v, extra: '%v'\n", chanReq.ChannelType(), hex.EncodeToString(chanReq.ExtraData()))
	switch chanReq.ChannelType() {
	case "session":
		if ch, reqs, err := chanReq.Accept(); err != nil {
			fmt.Fprintf(debugStream, "fail to accept channel request: %v\n", err)
			chanReq.Reject(ssh.ResourceShortage, "channel accept failure")
		} else {
			chsvr := &sshSessionChannelServer{
				sshChannelServer: &sshChannelServer{svr, chanReq, ch, reqs},
				env:              append([]string{}, os.Environ()...),
			}
			chsvr.handle(debugStream)
		}
	default:
		chanReq.Reject(ssh.UnknownChannelType, "channel type is not a session")
	}
}

func (chsvr *sshSessionChannelServer) handle(debugStream io.Writer) {
	// should maybe do something here...
	go chsvr.handleReqs(debugStream)
}

func (chsvr *sshSessionChannelServer) handleReqs(debugStream io.Writer) {
	for req := range chsvr.newReqs {
		chsvr.handleReq(req, debugStream)
	}
	fmt.Fprintf(debugStream, "ssh server session channel complete\n")
}

func (chsvr *sshSessionChannelServer) handleReq(req *ssh.Request, debugStream io.Writer) {
	switch req.Type {
	case "env":
		chsvr.handleEnv(req, debugStream)
	case "subsystem":
		chsvr.handleSubsystem(req, debugStream)
	default:
		rejectRequest(req, debugStream)
	}
}

func rejectRequest(req *ssh.Request, debugStream io.Writer) error {
	fmt.Fprintf(debugStream, "ssh rejecting request, type: %s\n", req.Type)
	err := req.Reply(false, []byte{})
	if err != nil {
		fmt.Fprintf(debugStream, "ssh request reply had error: %v\n", err)
	}
	return err
}

func rejectRequestUnmarshalError(req *ssh.Request, s interface{}, err error, debugStream io.Writer) error {
	fmt.Fprintf(debugStream, "ssh request unmarshaling error, type '%T': %v\n", s, err)
	rejectRequest(req, debugStream)
	return err
}

// env request form:
type sshEnvRequest struct {
	Envvar string
	Value  string
}

func (chsvr *sshSessionChannelServer) handleEnv(req *ssh.Request, debugStream io.Writer) error {
	envReq := &sshEnvRequest{}
	if err := ssh.Unmarshal(req.Payload, envReq); err != nil {
		return rejectRequestUnmarshalError(req, envReq, err, debugStream)
	}
	req.Reply(true, nil)

	found := false
	for i, envstr := range chsvr.env {
		if strings.HasPrefix(envstr, envReq.Envvar+"=") {
			found = true
			chsvr.env[i] = envReq.Envvar + "=" + envReq.Value
		}
	}
	if !found {
		chsvr.env = append(chsvr.env, envReq.Envvar+"="+envReq.Value)
	}

	return nil
}

// Payload: int: command size, string: command
type sshSubsystemRequest struct {
	Name string
}

type sshSubsystemExitStatus struct {
	Status uint32
}

func (chsvr *sshSessionChannelServer) handleSubsystem(req *ssh.Request, debugStream io.Writer) error {
	defer func() {
		err1 := chsvr.ch.CloseWrite()
		err2 := chsvr.ch.Close()
		fmt.Fprintf(debugStream, "ssh server subsystem request complete, err: %v %v\n", err1, err2)
	}()

	subsystemReq := &sshSubsystemRequest{}
	if err := ssh.Unmarshal(req.Payload, subsystemReq); err != nil {
		return rejectRequestUnmarshalError(req, subsystemReq, err, debugStream)
	}

	// reply to the ssh client

	// no idea if this is actually correct spec-wise.
	// just enough for an sftp server to start.
	if subsystemReq.Name != "sftp" {
		return req.Reply(false, nil)
	}

	req.Reply(true, nil)

	sftpServer, err := sftp.NewServer(
		chsvr.ch,
		sftp.WithDebug(debugStream),
	)
	if err != nil {
		return err
	}

	// wait for the session to close
	runErr := sftpServer.Serve()
	exitStatus := uint32(1)
	if runErr == nil {
		exitStatus = uint32(0)
	}

	_, exitStatusErr := chsvr.ch.SendRequest("exit-status", false, ssh.Marshal(sshSubsystemExitStatus{exitStatus}))
	return exitStatusErr
}
