package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const network = "tcp"
const address = "0.0.0.0:5432"

// Key user for authentication
const PGUSER_KEY = "user"

// Value for key user
const PGUSER_VALUE = "postgres"

// protocol version for communication betweeb backend and frontend
const protocolV = 196608

func main() {
	conn := connect()
	defer conn.Close()

	msg := buildStartUpMessage(PGUSER_KEY, PGUSER_VALUE)
	conn.Write(msg)

	r := bufio.NewReader(conn)
	receive(r)

	sc := bufio.NewScanner(os.Stdin)
	fmt.Printf("-> ")
	for sc.Scan() {
		query := sc.Text()
		q := buildQueryMessage(query)
		conn.Write(q)
		queryResponse(r)
		fmt.Printf("->")
	}
}

// connect via Dial
func connect() net.Conn {
	conn, err := net.Dial(network, address)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

// buildStartUpMessage creates buffer
// and fills it with user creds, protocol and message len
func buildStartUpMessage(k, v string) []byte {
	protocol := make([]uint8, 4)
	binary.BigEndian.PutUint32(protocol, uint32(protocolV))

	msg := make([]uint8, 4)
	msg = append(msg, protocol...)

	msg = append(msg, []uint8("user\x00postgres\x00")...)
	msg = append(msg, '\x00')

	binary.BigEndian.PutUint32(msg, uint32(len(msg)))
	return msg
}

// buildQueryMessage creates buffer
// and fills it with given query
func buildQueryMessage(q string) []byte {
	tag := 'Q'

	meta := make([]byte, 4)
	payload := []byte(q)
	payload = append(payload, '\x00')
	payload = append(meta, payload...)
	binary.BigEndian.PutUint32(payload, uint32(len(payload)))

	payload = append([]byte{byte(tag)}, payload...)
	return payload
}

// receive reads stream from backend
func receive(r *bufio.Reader) {
	tag, _ := r.ReadByte()
	for tag != 90 {
		readMsg(r)
		tag, _ = r.ReadByte()
	}
	readMsg(r)
}

// readMsg reads message part after tag
func readMsg(r *bufio.Reader) []byte {
	n := readMsgLen(4, r)
	msg := make([]byte, n)
	io.ReadFull(r, msg)
	return msg
}

// readMsgLen reads len of response message from backend
func readMsgLen(n int, r *bufio.Reader) int {
	lenPart := make([]byte, n)
	io.ReadFull(r, lenPart)
	return int(binary.BigEndian.Uint32(lenPart)) - 4
}