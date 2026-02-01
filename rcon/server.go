package rcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"jjmc/auth"
	"jjmc/instances"
)

const (
	SERVERDATA_AUTH           = 3
	SERVERDATA_AUTH_RESPONSE  = 2
	SERVERDATA_EXECCOMMAND    = 2
	SERVERDATA_RESPONSE_VALUE = 0
)

type RCONServer struct {
	Addr            string
	AuthManager     *auth.AuthManager
	InstanceManager *instances.InstanceManager
	listener        net.Listener
}

func NewRCONServer(addr string, am *auth.AuthManager, im *instances.InstanceManager) *RCONServer {
	return &RCONServer{
		Addr:            addr,
		AuthManager:     am,
		InstanceManager: im,
	}
}

func (s *RCONServer) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	log.Printf("RCON Server listening on %s", s.Addr)

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

func (s *RCONServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *RCONServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	authenticated := false
	var instance *instances.Instance

	for {

		var size int32
		if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
			return
		}

		if size < 10 || size > 4096 {
			return
		}

		payload := make([]byte, size)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return
		}

		id := int32(binary.LittleEndian.Uint32(payload[0:4]))
		typ := int32(binary.LittleEndian.Uint32(payload[4:8]))
		body := payload[8 : len(payload)-2]

		switch typ {
		case SERVERDATA_AUTH:

			parts := strings.SplitN(string(body), "#", 2)
			password := parts[0]

			if s.AuthManager.VerifyPassword(password) {
				authenticated = true
				if len(parts) > 1 {
					instanceID := parts[1]
					if inst, err := s.InstanceManager.GetInstance(instanceID); err == nil {
						instance = inst
					}
				}

				sendBoxPacket(conn, id, SERVERDATA_AUTH_RESPONSE, "")
			} else {

				sendBoxPacket(conn, -1, SERVERDATA_AUTH_RESPONSE, "")
				return
			}

		case SERVERDATA_EXECCOMMAND:
			if !authenticated {
				sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, "Not authenticated")
				continue
			}
			if instance == nil {
				sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, "No instance selected (use password#instanceID)")
				continue
			}

			cmd := string(body)

			output, err := instance.Manager.ExecuteCommand(cmd, 100*time.Millisecond)
			if err != nil {
				sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, fmt.Sprintf("Error: %v", err))
			} else {
				sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, output)
			}

		default:
			sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, "Unknown type")
		}
	}
}

func sendBoxPacket(conn net.Conn, id int32, typ int32, body string) {
	buf := new(bytes.Buffer)

	size := int32(4 + 4 + len(body) + 2)

	binary.Write(buf, binary.LittleEndian, size)
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, typ)
	buf.WriteString(body)
	buf.WriteByte(0)
	buf.WriteByte(0)

	conn.Write(buf.Bytes())
}
