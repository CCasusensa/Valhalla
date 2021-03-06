package connection

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/crypt"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// Client -
type Client struct {
	net.Conn
	readingHeader bool
	cSend         crypt.Maple
	cRecv         crypt.Maple
}

// NewClient -
func NewClient(conn net.Conn) Client {
	client := Client{Conn: conn, readingHeader: true}

	key := [4]byte{}
	rand.Read(key[:])

	client.cSend = crypt.New(key, consts.MapleVersion)

	rand.Read(key[:])
	client.cRecv = crypt.New(key, consts.MapleVersion)

	err := sendHandshake(client)

	if err != nil {
		client.Close()
	}

	return client
}

// String -
func (handle Client) String() string {
	return fmt.Sprintf("[Address] %s", handle.Conn.RemoteAddr())
}

// Close -
func (handle *Client) Close() error {
	return handle.Conn.Close()
}

func (handle *Client) sendPacket(p maplepacket.Packet) error {
	_, err := handle.Conn.Write(p)
	return err
}

func (handle *Client) Write(p maplepacket.Packet) error {
	local := make([]byte, len(p))
	copy(local, p)

	handle.cSend.Encrypt(local, true, false)

	_, err := handle.Conn.Write(local)

	return err
}

func (handle *Client) Read(p maplepacket.Packet) error {
	_, err := handle.Conn.Read(p)

	if err != nil {
		return err
	}
	if handle.readingHeader == true {
		handle.readingHeader = false
	} else {
		handle.readingHeader = true
		handle.cRecv.Decrypt(p, true, false)
	}

	return err
}

func (handle *Client) GetClientIPPort() net.Addr {
	return handle.Conn.RemoteAddr()
}

func sendHandshake(client Client) error {
	packet := maplepacket.NewPacket()

	packet.WriteInt16(13)
	packet.WriteInt16(consts.MapleVersion)
	packet.WriteString("")
	packet.Append(client.cRecv.IV()[:4])
	packet.Append(client.cSend.IV()[:4])
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}
