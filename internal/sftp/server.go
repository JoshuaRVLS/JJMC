package sftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"jjmc/internal/auth"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPServer struct {
	Addr        string
	BasePath    string
	AuthManager *auth.AuthManager
	listener    net.Listener
}

func NewSFTPServer(addr string, basePath string, am *auth.AuthManager) *SFTPServer {
	return &SFTPServer{
		Addr:        addr,
		BasePath:    basePath,
		AuthManager: am,
	}
}

func (s *SFTPServer) Start() error {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if s.AuthManager.VerifyPassword(string(pass)) {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateKey, err := generatePrivateKey()
	if err != nil {
		return err
	}
	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("SFTP Server listening on %s", s.Addr)

	go func() {
		for {
			nConn, err := s.listener.Accept()
			if err != nil {
				return
			}

			go s.handleConnection(nConn, config)
		}
	}()

	return nil
}

func (s *SFTPServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *SFTPServer) handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "subsystem":
					if string(req.Payload[4:]) == "sftp" {
						ok = true
						server, err := sftp.NewServer(
							channel,
							sftp.WithServerWorkingDirectory(s.BasePath),
						)
						if err != nil {
							return
						}
						if err := server.Serve(); err == io.EOF {
							server.Close()
						}
					}
				}
				req.Reply(ok, nil)
			}
		}(requests)
	}
}

func generatePrivateKey() (ssh.Signer, error) {

	keyPath := "ssh_host_rsa_key"
	if _, err := os.Stat(keyPath); err == nil {
		bytes, err := os.ReadFile(keyPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(bytes)
			if err == nil {
				return signer, nil
			}
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	file, err := os.Create(keyPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := pem.Encode(file, pemBlock); err != nil {
		return nil, err
	}

	return ssh.NewSignerFromKey(key)
}
