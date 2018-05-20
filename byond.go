package byond

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type QueryClient struct {
	Host string
}

func (c *QueryClient) Query(ctx context.Context, query []byte, fetchResp bool) ([]byte, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.Host)
	if err != nil {
		return nil, err
	}

	if err := sendMsg(conn, query); err != nil {
		return nil, err
	}

	if !fetchResp {
		return nil, nil
	}

	resp, err := recvMsg(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendMsg(conn net.Conn, query []byte) error {
	buf := make([]byte, 11+len(query))

	buf[1] = 0x83
	binary.BigEndian.PutUint16(buf[2:4], uint16(len(query)+7))

	buf[9] = '?'
	for i, b := range []byte(query) {
		buf[i+10] = b
	}

	_, err := conn.Write(buf)
	return err
}

func recvMsg(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 5)
	if n, err := conn.Read(buf); n != 5 {
		return nil, fmt.Errorf("expected 5 bytes, got %d", n)
	} else if err != nil {
		return nil, err
	}

	if buf[0] != 0x00 || buf[1] != 0x83 || buf[4] != 0x06 {
		return nil, errors.New("bad response")
	}

	length := int(binary.BigEndian.Uint16(buf[2:4])) - 1

	buf = make([]byte, length)
	if n, err := conn.Read(buf); n != length {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	} else if err != nil {
		return nil, err
	}

	return buf[:length-1], nil
}

func NewQueryClient(host string) *QueryClient {
	return &QueryClient{
		Host: host,
	}
}
