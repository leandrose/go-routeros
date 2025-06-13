package go_routeros

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 4096)
		return &buf
	},
}

// readWord reads a word (based on the length prefix) from the reader
func (c *Client) readWord() (string, error) {
	length, err := c.decodeLength()
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}

	raw := bufPool.Get().(*[]byte)
	defer bufPool.Put(raw)

	buf := *raw
	if len(buf) < length {
		buf = make([]byte, length)
	}

	if _, err := io.ReadFull(c.reader, buf[:length]); err != nil {
		return "", err
	}
	return string(buf[:length]), nil
}

// readSentence reads a sentence (list of words) from the connection
func (c *Client) readSentence() (map[string]string, error) {
	sentence := make(map[string]string)
	for {
		word, err := c.readWord()
		if c.debug {
			fmt.Printf("DEBUG READER: %s\n", word)
		}
		if err != nil {
			return nil, err
		}
		if word == "" {
			break
		}
		if strings.HasPrefix(word, "!") {
			sentence["!type"] = word
		} else if strings.Contains(word, "=") {
			parts := strings.SplitN(strings.TrimLeft(word, "="), "=", 2)
			sentence[parts[0]] = parts[1]
		}
	}
	return sentence, nil
}
