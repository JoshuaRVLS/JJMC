package telnet

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"jjmc/auth"
	"jjmc/servers"
)

type TelnetServer struct {
	Addr            string
	AuthManager     *auth.AuthManager
	InstanceManager *servers.InstanceManager
}

func NewTelnetServer(addr string, am *auth.AuthManager, im *servers.InstanceManager) *TelnetServer {
	return &TelnetServer{
		Addr:            addr,
		AuthManager:     am,
		InstanceManager: im,
	}
}

func (s *TelnetServer) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	log.Printf("Telnet Server listening on %s", s.Addr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Telnet Accept Error: %v", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

func (s *TelnetServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Banner
	rw.WriteString("JJMC Telnet Console\r\n")
	rw.WriteString("====================\r\n")
	rw.Flush()

	// Auth
	// Since we only have a global password, just ask for "Password:"
	// Or mimicking PufferPanel: Username/Password?
	// Our AuthManager only supports single password for now.
	rw.WriteString("Password: ")
	rw.Flush()

	pass, err := rw.Reader.ReadString('\n')
	if err != nil {
		return
	}
	pass = strings.TrimSpace(pass)

	if !s.AuthManager.VerifyPassword(pass) {
		rw.WriteString("Authentication Failed.\r\n")
		rw.Flush()
		return
	}

	rw.WriteString("Authenticated.\r\n")

	// Select Instance
	rw.WriteString("Enter Instance ID: ")
	rw.Flush()

	id, err := rw.Reader.ReadString('\n')
	if err != nil {
		return
	}
	id = strings.TrimSpace(id)

	inst, err := s.InstanceManager.GetInstance(id)
	if err != nil {
		rw.WriteString("Instance not found.\r\n")
		rw.Flush()
		return
	}

	rw.WriteString(fmt.Sprintf("Attached to %s (%s). Type /exit to disconnect.\r\n", inst.Name, inst.ID))
	rw.Flush()

	// Attach Client
	client := &TelnetClient{conn: conn}
	inst.Manager.RegisterClient(client)
	defer inst.Manager.UnregisterClient(client)

	// Read Input Loop
	for {
		line, err := rw.Reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		if line == "/exit" {
			break
		}

		if line != "" {
			inst.Manager.WriteCommand(line)
		}
	}
}

type TelnetClient struct {
	conn net.Conn
}

func (c *TelnetClient) WriteMessage(messageType int, data []byte) error {
	// messageType is ignored (assumed text)
	// Telnet expects CRLF?
	_, err := c.conn.Write(append(data, '\r', '\n'))
	return err
}
