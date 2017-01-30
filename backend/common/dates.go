package common

import (
	"errors"
	"fmt"
	"time"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// DateOnlyTime provides a time.Time construct with a 24 Hour
// precision after parsing.
type DateOnlyTime struct {
	time.Time
}

func (dot *DateOnlyTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if len(s) == 2 {
		return errors.New("Cannot parse empty date")
	}
	s = s[1 : len(s)-1]

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	dot.Time = t.UTC()
	return nil
}

func (dot *DateOnlyTime) MarshalJSON() ([]byte, error) {
	s := dot.Format("2006-01-02")
	s = fmt.Sprintf("\"%s\"", s)
	return []byte(s), nil
}

func (dot *DateOnlyTime) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeInt64(dot.Unix())
}

func (dot *DateOnlyTime) DecodeMsgpack(dec *msgpack.Decoder) error {
	dotunix, err := dec.DecodeInt64()
	if err != nil {
		return err
	}
	dot.Time = time.Unix(dotunix, 0).UTC()
	return nil
}

func (dot *DateOnlyTime) TTL() time.Duration {
	dead := dot.Unix()
	now := time.Now().UTC().Unix()
	// flake already dead, TTL is 0
	if now >= dead {
		return 0 * time.Second
	}
	return time.Duration(dead-now) * time.Second
}

func FromTime(t time.Time) *DateOnlyTime {
	return &DateOnlyTime{
		Time: t.UTC().Truncate(24 * time.Hour),
	}
}

func FromString(s string) (*DateOnlyTime, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return FromTime(t), nil
}
