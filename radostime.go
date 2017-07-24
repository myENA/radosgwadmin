package radosgwadmin

import (
	"errors"
	"time"
)

// RadosTime - This knows how to use the date time formats returned
// from the rados gateway.
type RadosTime time.Time

// RadosTimeFormat - this is the most common rados time format
const RadosTimeFormat string = "2006-01-02 15:04:05.000000Z07:00"

// RadosBucketTimeFormat - used for bucket calls
const RadosBucketTimeFormat string = "2006-01-02 15:04:05.000000"

// UnmarshalText - implements TextUnmarshaler
func (rt *RadosTime) UnmarshalText(text []byte) error {
	var err error
	t := (*time.Time)(rt)
	*t, err = time.Parse(RadosTimeFormat, string(text))
	if err != nil {
		*t, err = time.ParseInLocation(RadosBucketTimeFormat, string(text), tz)
	}
	return err
}

// MarshalText - implements TextMarshaler
func (rt RadosTime) MarshalText() ([]byte, error) {
	t := time.Time(rt)
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("RadosTime.MarshalText(): year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(RadosTimeFormat))
	return t.AppendFormat(b, RadosTimeFormat), nil
}
