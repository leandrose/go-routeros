package go_routeros

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	conn        io.ReadWriteCloser
	lock        sync.Mutex
	reader      *bufio.Reader
	responses   map[int]chan Response
	nextID      int
	debug       bool
	loopMutex   sync.Mutex
	isConnected bool
}

// Dial dial
func Dial(addr string) (*Client, error) {
	return DialContext(context.Background(), addr)
}

// DialTimeout dial with context and timeout
func DialTimeout(duration time.Duration, addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	return DialContext(ctx, addr)
}

// DialContext dial with context
func DialContext(ctx context.Context, addr string) (*Client, error) {
	conn, err := (new(net.Dialer)).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to router os: %w", err)
	}
	return &Client{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		responses:   make(map[int]chan Response),
		isConnected: false,
	}, nil
}

// DialTLS dial
func DialTLS(address string, tlsConfig *tls.Config) (*Client, error) {
	return DialTLSContext(context.Background(), address, tlsConfig)
}

// DialTLSTimeout dial with context and timeout
func DialTLSTimeout(address string, tlsConfig *tls.Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return DialTLSContext(ctx, address, tlsConfig)
}

// DialTLSContext dial with context
func DialTLSContext(ctx context.Context, address string, tlsConfig *tls.Config) (*Client, error) {
	conn, err := (&tls.Dialer{Config: tlsConfig}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("could not connect to router os: %w", err)
	}
	return &Client{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		responses:   make(map[int]chan Response),
		isConnected: false,
	}, nil
}

// Close the connection
func (c *Client) Close() {
	c.isConnected = false
	if c.conn != nil {
		_ = c.conn.Close()
	}
	for k, ch := range c.responses {
		c.lock.Lock()
		delete(c.responses, k)
		c.lock.Unlock()
		close(ch)
	}
}

// Login to routeros
func (c *Client) Login(username, password string) error {
	sent := []string{
		"/login",
		fmt.Sprintf("=name=%s", username),
		fmt.Sprintf("=password=%s", password),
	}
	if err := c.writeSentence(sent); err != nil {
		return err
	}

	for {
		sentence, err := c.readSentence()
		if c.debug {
			fmt.Printf("DEBUG LOGIN: %+v\n", sentence)
		}
		if err != nil {
			return err
		}
		if v, ok := sentence["!type"]; ok {
			if v == "!done" {
				c.isConnected = true
				go c.readLoop()
				return nil
			}
			if sentence["!type"] == "!trap" || sentence["!type"] == "!fatal" {
				return errors.New(sentence["=message"])
			}
		}
	}
}

// IsConnected check if the client is connected
func (c *Client) IsConnected() bool {
	return c.isConnected
}

// SendCommand sends a command to RouterOS and returns a channel to receive responses
func (c *Client) SendCommand(cmd string, args ...string) (chan Response, error) {
	c.lock.Lock()
	id := c.nextID
	c.nextID++
	ch := make(chan Response, 10)
	c.responses[id] = ch
	c.lock.Unlock()

	fullCmd := []string{cmd}
	fullCmd = append(fullCmd, args...)
	fullCmd = append(fullCmd, fmt.Sprintf(".tag=%d", id))

	err := c.writeSentence(fullCmd)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// sendErrorAllResponses send an error response to all channels
func (c *Client) sendErrorAllResponses(err error) {
	for _, ch := range c.responses {
		ch <- Response{
			Err: &RouterOSError{
				message: err.Error(),
			},
			Type: "!fatal",
			Data: map[string]string{
				"message": err.Error(),
			},
		}
	}
}

func (c *Client) readLoop() {
	if !c.loopMutex.TryLock() {
		return
	}
	defer c.loopMutex.Unlock()
	for {
		sentence, err := c.readSentence()
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.sendErrorAllResponses(err)
				c.Close()
				return
			}
			continue
		}

		tag := sentence[".tag"]
		if tag == "" {
			continue
		}

		id, err := strconv.ParseInt(tag, 10, 32)
		if err != nil {
			continue
		}

		c.lock.Lock()
		ch, ok := c.responses[int(id)]
		c.lock.Unlock()

		if !ok {
			continue
		}

		response := Response{
			Type: sentence["!type"],
			Data: sentence,
		}

		if response.Type == "!trap" || response.Type == "!fatal" {
			errRouterOS := RouterOSError{}
			if e, ok := response.Data["message"]; ok {
				errRouterOS.message = e
			} else {
				errRouterOS.message = "an error occurred"
			}
			response.Err = &errRouterOS
		}

		ch <- response

		if response.Type == "!done" || response.Type == "!trap" || response.Type == "!fatal" || response.Type == "!empty" {
			c.lock.Lock()
			delete(c.responses, int(id))
			c.lock.Unlock()
			close(ch)
		}
	}
}

func (c *Client) EnableDebug() {
	c.debug = true
}

func (c *Client) DisableDebug() {
	c.debug = false
}
