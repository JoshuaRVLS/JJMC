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

	"jjmc/auth"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPServer struct {
	Addr        string
	BasePath    string
	AuthManager *auth.AuthManager
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
			// Use existing AuthManager logic
			if s.AuthManager.VerifyPassword(string(pass)) {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	// Generate private key for the server
	// In production, persist this key!
	privateKey, err := generatePrivateKey()
	if err != nil {
		return err
	}
	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	log.Printf("SFTP Server listening on %s", s.Addr)

	go func() {
		for {
			nConn, err := listener.Accept()
			if err != nil {
				log.Printf("SFTP Accept error: %v", err)
				continue
			}

			go s.handleConnection(nConn, config)
		}
	}()

	return nil
}

func (s *SFTPServer) handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	// Before using, shake hands
	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		// log.Printf("SFTP Handshake failed: %v", err)
		return
	}

	// Servicing the incoming requests
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

		// Sessions have out-of-band requests such as "subsystem" => "sftp"
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "subsystem":
					if string(req.Payload[4:]) == "sftp" { // Payload usually has length prefix
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

// Helper to generate an ephemeral host key
// In a real app, load from disk
func generatePrivateKey() (ssh.Signer, error) {
	// For simplicity, generate RSA 2048
	// This is slow on startup, better to save it.
	// But sufficient for demo/dev.
	// Or check if "ssh_host_rsa_key" exists.

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

	// Generate new (stub logic for now, using a hardcoded dummy or generating via crypto/rsa if crucial)
	// Implementing actual generation requires "crypto/rsa" and "crypto/rand"
	// To avoid complexity, let's skip generation code here and return error if not found,
	// OR (better) use a pre-generated one or generate on fly.
	// For this task, I'll allow generating on fly for completeness.

	// Generate new key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Save to disk
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
