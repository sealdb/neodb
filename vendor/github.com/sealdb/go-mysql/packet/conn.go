package packet

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"sync"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"

	"github.com/juju/errors"
	. "github.com/sealdb/go-mysql/mysql"
)

type BufPool struct {
	pool *sync.Pool
}

func NewBufPool() *BufPool {
	return &BufPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (b *BufPool) Get() *bytes.Buffer {
	return b.pool.Get().(*bytes.Buffer)
}

func (b *BufPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	b.pool.Put(buf)
}

/*
	Conn is the base class to handle MySQL protocol.
*/
type Conn struct {
	net.Conn

	// we removed the buffer reader because it will cause the SSLRequest to block (tls connection handshake won't be
	// able to read the "Client Hello" data since it has been buffered into the buffer reader)

	bufPool *BufPool
	br      *bufio.Reader
	reader  io.Reader

	Sequence uint8
}

func NewConn(conn net.Conn) *Conn {
	c := new(Conn)
	c.Conn = conn

	c.bufPool = NewBufPool()
	c.br = bufio.NewReaderSize(c, 65536) // 64kb
	c.reader = c.br

	return c
}

func NewTLSConn(conn net.Conn) *Conn {
	c := new(Conn)
	c.Conn = conn

	c.bufPool = NewBufPool()
	c.reader = c

	return c
}

func (c *Conn) ReadPacket() ([]byte, error) {
	// Here we use `sync.Pool` to avoid allocate/destroy buffers frequently.
	buf := c.bufPool.Get()
	defer c.bufPool.Put(buf)

	if err := c.ReadPacketTo(buf); err != nil {
		return nil, errors.Trace(err)
	} else {
		result := append([]byte{}, buf.Bytes()...)
		return result, nil
	}
}

func (c *Conn) ReadPacketTo(w io.Writer) error {
	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(c.reader, header); err != nil {
		return ErrBadConn
	}

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	sequence := uint8(header[3])

	if sequence != c.Sequence {
		return errors.Errorf("invalid sequence %d != %d", sequence, c.Sequence)
	}

	c.Sequence++

	if buf, ok := w.(*bytes.Buffer); ok {
		// Allocate the buffer with expected length directly instead of call `grow` and migrate data many times.
		buf.Grow(length)
	}

	if n, err := io.CopyN(w, c.reader, int64(length)); err != nil {
		return ErrBadConn
	} else if n != int64(length) {
		return ErrBadConn
	} else {
		if length < MaxPayloadLen {
			return nil
		}

		if err := c.ReadPacketTo(w); err != nil {
			return err
		}
	}

	return nil
}

// WritePacket: data already has 4 bytes header
// will modify data inplace
func (c *Conn) WritePacket(data []byte) error {
	length := len(data) - 4

	for length >= MaxPayloadLen {
		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = c.Sequence

		if n, err := c.Write(data[:4+MaxPayloadLen]); err != nil {
			return ErrBadConn
		} else if n != (4 + MaxPayloadLen) {
			return ErrBadConn
		} else {
			c.Sequence++
			length -= MaxPayloadLen
			data = data[MaxPayloadLen:]
		}
	}

	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = c.Sequence

	if n, err := c.Write(data); err != nil {
		return ErrBadConn
	} else if n != len(data) {
		return ErrBadConn
	} else {
		c.Sequence++
		return nil
	}
}

// WriteClearAuthPacket: Client clear text authentication packet
// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::AuthSwitchResponse
func (c *Conn) WriteClearAuthPacket(password string) error {
	// Calculate the packet length and add a tailing 0
	pktLen := len(password) + 1
	data := make([]byte, 4+pktLen)

	// Add the clear password [null terminated string]
	copy(data[4:], password)
	data[4+pktLen-1] = 0x00

	return c.WritePacket(data)
}

// WritePublicKeyAuthPacket: Caching sha2 authentication. Public key request and send encrypted password
// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::AuthSwitchResponse
func (c *Conn) WritePublicKeyAuthPacket(password string, cipher []byte) error {
	// request public key
	data := make([]byte, 4+1)
	data[4] = 2 // cachingSha2PasswordRequestPublicKey
	c.WritePacket(data)

	data, err := c.ReadPacket()
	if err != nil {
		return err
	}

	block, _ := pem.Decode(data[1:])
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	plain := make([]byte, len(password)+1)
	copy(plain, password)
	for i := range plain {
		j := i % len(cipher)
		plain[i] ^= cipher[j]
	}
	sha1v := sha1.New()
	enc, _ := rsa.EncryptOAEP(sha1v, rand.Reader, pub.(*rsa.PublicKey), plain, nil)
	data = make([]byte, 4+len(enc))
	copy(data[4:], enc)
	return c.WritePacket(data)
}

func (c *Conn) WriteEncryptedPassword(password string, seed []byte, pub *rsa.PublicKey) error {
	enc, err := EncryptPassword(password, seed, pub)
	if err != nil {
		return err
	}
	return c.WriteAuthSwitchPacket(enc, false)
}

// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::AuthSwitchResponse
func (c *Conn) WriteAuthSwitchPacket(authData []byte, addNUL bool) error {
	pktLen := 4 + len(authData)
	if addNUL {
		pktLen++
	}
	data := make([]byte, pktLen)

	// Add the auth data [EOF]
	copy(data[4:], authData)
	if addNUL {
		data[pktLen-1] = 0x00
	}

	return c.WritePacket(data)
}

func (c *Conn) ResetSequence() {
	c.Sequence = 0
}

func (c *Conn) Close() error {
	c.Sequence = 0
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}
