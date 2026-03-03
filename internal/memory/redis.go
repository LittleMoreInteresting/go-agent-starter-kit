package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type RedisStore struct {
	addr      string
	password  string
	db        int
	keyPrefix string
}

func NewRedisStore(addr, password, keyPrefix string, db int) *RedisStore {
	return &RedisStore{addr: addr, password: password, keyPrefix: keyPrefix, db: db}
}

func (s *RedisStore) key(sessionID string) string {
	return fmt.Sprintf("%s:%s", s.keyPrefix, sessionID)
}

func (s *RedisStore) Append(ctx context.Context, sessionID string, msg Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	_, err = s.exec(ctx, "RPUSH", s.key(sessionID), string(b))
	return err
}

func (s *RedisStore) History(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	start, end := "0", "-1"
	if limit > 0 {
		start = strconv.Itoa(-limit)
	}
	raw, err := s.exec(ctx, "LRANGE", s.key(sessionID), start, end)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(raw, "\n")
	msgs := make([]Message, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m Message
		if json.Unmarshal([]byte(line), &m) == nil {
			msgs = append(msgs, m)
		}
	}
	return msgs, nil
}

func (s *RedisStore) Close() error { return nil }

func (s *RedisStore) exec(ctx context.Context, args ...string) (string, error) {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return "", fmt.Errorf("redis dial: %w", err)
	}
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	if s.password != "" {
		if err := writeRESP(rw, "AUTH", s.password); err != nil {
			return "", err
		}
		if _, err := readRESP(rw); err != nil {
			return "", err
		}
	}
	if s.db > 0 {
		if err := writeRESP(rw, "SELECT", strconv.Itoa(s.db)); err != nil {
			return "", err
		}
		if _, err := readRESP(rw); err != nil {
			return "", err
		}
	}
	if err := writeRESP(rw, args...); err != nil {
		return "", err
	}
	return readRESP(rw)
}

func writeRESP(rw *bufio.ReadWriter, args ...string) error {
	if _, err := rw.WriteString("*" + strconv.Itoa(len(args)) + "\r\n"); err != nil {
		return err
	}
	for _, a := range args {
		if _, err := rw.WriteString("$" + strconv.Itoa(len(a)) + "\r\n" + a + "\r\n"); err != nil {
			return err
		}
	}
	return rw.Flush()
}

func readRESP(rw *bufio.ReadWriter) (string, error) {
	prefix, err := rw.ReadByte()
	if err != nil {
		return "", err
	}
	line, err := rw.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\r\n")
	switch prefix {
	case '+', ':':
		return line, nil
	case '-':
		return "", fmt.Errorf("redis error: %s", line)
	case '$':
		n, _ := strconv.Atoi(line)
		if n < 0 {
			return "", nil
		}
		buf := make([]byte, n+2)
		if _, err := rw.Read(buf); err != nil {
			return "", err
		}
		return string(buf[:n]), nil
	case '*':
		n, _ := strconv.Atoi(line)
		items := make([]string, 0, n)
		for i := 0; i < n; i++ {
			v, err := readRESP(rw)
			if err != nil {
				return "", err
			}
			items = append(items, v)
		}
		return strings.Join(items, "\n"), nil
	default:
		return "", fmt.Errorf("unknown redis resp prefix: %q", prefix)
	}
}
