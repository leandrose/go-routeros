package go_routeros

import (
    "encoding/binary"
    "errors"
)

// encodeLength encodes the word length in Mikrotik format
func encodeLength(length int) []byte {
    switch {
    case length < 0x80:
        return []byte{byte(length)}
    case length < 0x4000:
        l := uint16(length) | 0x8000
        return []byte{byte(l >> 8), byte(l)}
    case length < 0x200000:
        l := uint32(length) | 0xC00000
        return []byte{
            byte(l >> 16),
            byte(l >> 8),
            byte(l),
        }
    case length < 0x10000000:
        l := uint32(length) | 0xE0000000
        return []byte{
            byte(l >> 24),
            byte(l >> 16),
            byte(l >> 8),
            byte(l),
        }
    default:
        return []byte{
            0xF0,
            byte(length >> 24),
            byte(length >> 16),
            byte(length >> 8),
            byte(length),
        }
    }
}

// decodeLength reads the length prefix and returns the value (as an int)
func (c *Client) decodeLength() (int, error) {
    first, err := c.reader.ReadByte()
    if err != nil {
        return 0, err
    }

    switch {
    case first&0x80 == 0x00:
        return int(first), nil
    case first&0xC0 == 0x80:
        b, err := c.reader.Peek(1)
        if err != nil {
            return 0, err
        }
        _, _ = c.reader.Discard(1)
        return int(first&^0xC0)<<8 | int(b[0]), nil
    case first&0xE0 == 0xC0:
        b, err := c.reader.Peek(2)
        if err != nil {
            return 0, err
        }
        _, _ = c.reader.Discard(2)
        return int(first&^0xE0)<<16 | int(b[0])<<8 | int(b[1]), nil
    case first&0xF0 == 0xE0:
        b, err := c.reader.Peek(3)
        if err != nil {
            return 0, err
        }
        _, _ = c.reader.Discard(3)
        return int(first&^0xF0)<<24 | int(b[0])<<16 | int(b[1])<<8 | int(b[2]), nil
    case first == 0xF0:
        b, err := c.reader.Peek(4)
        if err != nil {
            return 0, err
        }
        _, _ = c.reader.Discard(4)
        return int(binary.BigEndian.Uint32(b)), nil
    default:
        return 0, errors.New("invalid length header")
    }
}
