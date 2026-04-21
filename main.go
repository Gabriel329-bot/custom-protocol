package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	gutls "github.com/refraction-networking/utls"
)

const (
	MagicHeader      = "AMNZ"
	ProtocolV1      = 0x01
	MaxPacketSize   = 65535
	HandshakeTimeout = 10 * time.Second
)

var (
	ErrInvalidMagic    = errors.New("invalid magic header")
	ErrInvalidVersion  = errors.New("invalid protocol version")
	ErrHandshake      = errors.New("handshake failed")
	ErrVerifyFailed  = errors.New("verification failed")
)

type Config struct {
	ListenAddr  string
	ServerName  string
	Fingerprint string
	Obfuscation uint8
	Compress    bool
	PrivateKey  string
	PublicKey   string
}

type Packet struct {
	Magic    [4]byte
	Version uint8
	Type    uint8
	Seq    uint32
	Length uint16
	Data   []byte
}

type Client struct {
	conn    net.Conn
	config *Config
	seq    uint32
	mu     sync.Mutex
}

type Server struct {
	conn    net.Listener
	config  *Config
	mu      sync.RWMutex
	clients map[net.Conn]*Client
}

func main() {
	listen := flag.String("l", ":443", "Listen address")
	serverName := flag.String("server-name", "www.microsoft.com", "SNI for TLS")
	fingerprint := flag.String("fingerprint", "Chrome", "TLS fingerprint")
	obfuscation := flag.Int("obfuscation", 3, "Obfuscation level (0-3)")
	mode := flag.String("mode", "server", "Mode: server or client")
	serverHost := flag.String("connect", "", "Server to connect to (client mode)")
	
	flag.Parse()

	cfg := &Config{
		ListenAddr:  *listen,
		ServerName: *serverName,
		Fingerprint: *fingerprint,
		Obfuscation: uint8(*obfuscation),
	}

	switch *mode {
	case "server":
		runServer(cfg)
	case "client":
		if *serverHost == "" {
			fmt.Println("Error: --connect required for client mode")
			os.Exit(1)
		}
		runClient(cfg, *serverHost)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Custom Protocol - AmneziaWG + TLS + Custom Handshake

Usage:
  customproto server -l :443                    Start server
  customproto client -connect server:443      Connect to server
  customproto keygen                     Generate keys

Options:
  -l :443                              Listen address
  -server-name www.microsoft.com         SNI for TLS
  -fingerprint Chrome                 TLS fingerprint (Chrome/Firefox/Edge/Safari)
  -obfuscation 3                       Obfuscation level (0-3)`)
}

func runServer(cfg *Config) {
	priv, pub := generateKeyPair()
	cfg.PrivateKey = hex.EncodeToString(priv[:])
	cfg.PublicKey = hex.EncodeToString(pub[:])

	fmt.Printf("Starting server on %s\n", cfg.ListenAddr)
	fmt.Printf("Public Key: %s\n", cfg.PublicKey)
	fmt.Printf("SNI: %s\n", cfg.ServerName)
	fmt.Printf("Fingerprint: %s\n", cfg.Fingerprint)

	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, cfg)
	}
}

func handleConnection(conn net.Conn, cfg *Config) {
	defer conn.Close()
	fmt.Printf("Client connected from %s\n", conn.RemoteAddr())

	client := &Client{conn: conn, config: cfg}

	for {
		packet, err := client.readPacket()
		if err != nil {
			break
		}

		if packet.Type == 0x01 {
			data := deobfuscate(packet.Data, cfg.Obfuscation)
			fmt.Printf("Received %d bytes\n", len(data))
			client.writePacket([]byte("ACK"))
		}
	}
}

func runClient(cfg *Config, serverAddr string) {
	priv, pub := generateKeyPair()
	cfg.PrivateKey = hex.EncodeToString(priv[:])
	cfg.PublicKey = hex.EncodeToString(pub[:])

	fmt.Printf("Connecting to %s\n", serverAddr)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	tlsConfig := &gutls.Config{
		ServerName:         cfg.ServerName,
		NextProtos:        []string{"h2", "http/1.1"},
		MinVersion:       gutls.VersionTLS13,
		MaxVersion:       gutls.VersionTLS13,
		InsecureSkipVerify: true,
	}

	uconn := gutls.UClient(conn, tlsConfig, gutls.ClientHelloID{
		Client: cfg.Fingerprint,
	})

	if err := uconn.Handshake(); err != nil {
		fmt.Println("TLS handshake error:", err)
		os.Exit(1)
	}

	fmt.Printf("Connected! TLS fingerprint: %s\n", cfg.Fingerprint)

	client := &Client{conn: conn, config: cfg}

	for i := 0; i < 5; i++ {
		test := []byte(fmt.Sprintf("Test message #%d", i+1))
		client.writePacket(test)
		fmt.Printf("Sent: %s\n", test)
		time.Sleep(2 * time.Second)
	}
}

func generateKeyPair() ([32]byte, [32]byte) {
	var priv [32]byte
	rand.Read(priv[:])
	hash := sha256.Sum256(priv[:])
	var pub [32]byte
	copy(pub[:], hash[:32])
	return priv, pub
}

func (c *Client) writePacket(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++

	obfuscated := obfuscate(data, c.config.Obfuscation)

	packet := Packet{
		Magic:    [4]byte{'A', 'M', 'N', 'Z'},
		Version: ProtocolV1,
		Type:    0x01,
		Seq:     c.seq,
		Length: uint16(len(obfuscated)),
		Data:   obfuscated,
	}

	size := 4 + 1 + 1 + 4 + 2 + packet.Length
	buf := make([]byte, size)

	copy(buf[0:4], packet.Magic[:])
	buf[4] = packet.Version
	buf[5] = packet.Type
	binary.BigEndian.PutUint32(buf[6:10], packet.Seq)
	binary.BigEndian.PutUint16(buf[10:12], packet.Length)
	copy(buf[12:12+packet.Length], packet.Data)

	_, err := c.conn.Write(buf)
	return err
}

func (c *Client) readPacket() (*Packet, error) {
	header := make([]byte, 12)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return nil, err
	}

	magic := [4]byte{}
	copy(magic[:], header[0:4])
	if string(magic[:]) != MagicHeader {
		return nil, ErrInvalidMagic
	}

	length := binary.BigEndian.Uint16(header[10:12])
	seq := binary.BigEndian.Uint32(header[6:10])

	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}

	return &Packet{
		Magic:   magic,
		Version: header[4],
		Type:    header[5],
		Seq:     seq,
		Length: length,
		Data:   data,
	}, nil
}

func obfuscate(data []byte, jc uint8) []byte {
	if jc == 0 {
		return data
	}

	junk := make([]byte, int(jc)*64)
	rand.Read(junk)

	result := make([]byte, 0, len(data)+len(junk))
	headerLen := int(jc) * 32
	result = append(result, junk[:headerLen]...)
	result = append(result, data...)
	result = append(result, junk[headerLen:]...)

	return result
}

func deobfuscate(data []byte, jc uint8) []byte {
	if jc == 0 || len(data) < 64 {
		return data
	}

	headerLen := int(jc) * 32
	trailerLen := int(jc) * 32

	if len(data) > headerLen+trailerLen {
		return data[headerLen : len(data)-trailerLen]
	}

	return data
}