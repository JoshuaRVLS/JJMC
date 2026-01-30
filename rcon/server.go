package rcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"jjmc/auth"
	"jjmc/servers"
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
	InstanceManager *servers.InstanceManager
}

func NewRCONServer(addr string, am *auth.AuthManager, im *servers.InstanceManager) *RCONServer {
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
	log.Printf("RCON Server listening on %s", s.Addr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("RCON Accept Error: %v", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}()
	return nil
}

func (s *RCONServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	authenticated := false
	var instance *servers.Instance

	for {
		// Read Packet
		// Size (4 bytes) little-endian
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

		// Parse
		id := int32(binary.LittleEndian.Uint32(payload[0:4]))
		typ := int32(binary.LittleEndian.Uint32(payload[4:8]))
		body := payload[8 : len(payload)-2] // Strip null terminators (2 bytes at end usually)

		// Handle
		switch typ {
		case SERVERDATA_AUTH:
			// Password format: password#instanceID
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
				// Auth Success Response: ID sent back
				sendBoxPacket(conn, id, SERVERDATA_AUTH_RESPONSE, "")
			} else {
				// Auth Fail: -1
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

			// Exec
			cmd := string(body)
			if err := instance.Manager.WriteCommand(cmd); err != nil {
				sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, fmt.Sprintf("Error: %v", err))
			} else {
				// RCON expects output. But our writes are async broadcasted.
				// Returning empty "Command sent" for now.
				// Ideally we capture next logs.
				sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, "")
			}

		default:
			sendBoxPacket(conn, id, SERVERDATA_RESPONSE_VALUE, "Unknown type")
		}
	}
}

func sendBoxPacket(conn net.Conn, id int32, typ int32, body string) {
	buf := new(bytes.Buffer)

	// Size = 4 (ID) + 4 (Type) + len(body) + 2 (Nulls)
	size := int32(4 + 4 + len(body) + 2)

	binary.Write(buf, binary.LittleEndian, size)
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, typ)
	buf.WriteString(body)
	buf.WriteByte(0)
	buf.WriteByte(0)

	conn.Write(buf.Bytes())
}
