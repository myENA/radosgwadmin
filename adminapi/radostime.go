package adminapi

import (
	"time"
	"errors"
)

type RadosTime time.Time

const RadosTimeFormat string = "2006-01-02 15:04:05.000000Z07:00"

func (rt *RadosTime) UnmarshalText(text []byte) error {
	var err error
	t := (*time.Time)(rt)
	*t, err = time.Parse(RadosTimeFormat, string(text))
	return err
}

func (rt RadosTime) MarshalText() ([]byte, error) {
	t := time.Time(rt)
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("RadosTime.MarshalText(): year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(RadosTimeFormat))
	return t.AppendFormat(b, RadosTimeFormat), nil
}

func (rt RadosTime) Time() time.Time {
	return time.Time(rt)
}
