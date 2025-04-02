package go_routeros

import (
    "fmt"
    "io"
    "strings"
)

// readWord reads a word (based on the length prefix) from the reader
func (c *Client) readWord() (string, error) {
    length, err := c.decodeLength()
    if err != nil {
        return "", err
    }
    if length == 0 {
        return "", nil
    }
    buf := make([]byte, length)
    if _, err := io.ReadFull(c.reader, buf); err != nil {
        return "", err
    }
    return string(buf), nil
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
