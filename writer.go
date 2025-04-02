package go_routeros

import "fmt"

func (c *Client) writeSentence(words []string) error {
    for _, word := range words {
        if c.debug {
            fmt.Printf("DEBUG WRITER: %s\n", word)
        }
        if err := c.writeWord(word); err != nil {
            return err
        }
    }
    return c.writeWord("") // fim da sentença
}

// writeWord writes a word (with length prefix) to the connection
func (c *Client) writeWord(word string) error {
    length := encodeLength(len(word))
    if _, err := c.conn.Write(length); err != nil {
        return err
    }
    if _, err := c.conn.Write([]byte(word)); err != nil {
        return err
    }
    return nil
}
