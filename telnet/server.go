package telnet

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"jjmc/auth"
	"jjmc/instances"
)

type TelnetServer struct {
	Addr            string
	AuthManager     *auth.AuthManager
	InstanceManager *instances.InstanceManager
	listener        net.Listener
}

func NewTelnetServer(addr string, am *auth.AuthManager, im *instances.InstanceManager) *TelnetServer {
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
	s.listener = listener
	log.Printf("Telnet Server listening on %s", s.Addr)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {

				return
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

func (s *TelnetServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *TelnetServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	rw.WriteString("JJMC Telnet Console\r\n")
	rw.WriteString("====================\r\n")
	rw.Flush()

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

	client := &TelnetClient{conn: conn}
	inst.Manager.RegisterClient(client)
	defer inst.Manager.UnregisterClient(client)

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

	_, err := c.conn.Write(append(data, '\r', '\n'))
	return err
}
