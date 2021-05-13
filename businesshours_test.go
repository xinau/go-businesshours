package businesshours

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestParseWeekday(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Weekday
		wantErr error
	}{{
		"sunday",
		"Sun",
		0,
		nil,
	}, {
		"saturday",
		"Sat",
		6,
		nil,
	}, {
		"invalid format",
		"invalid format",
		0,
		ErrorParseWeekday,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseWeekday(test.input)
			Assertf(t, errors.Is(err, test.wantErr), "got %v, expected %v", err, test.wantErr)
			Assertf(t, got == test.want, "got %d, expected %d", got, test.want)
		})
	}
}

func TestWeekday_String(t *testing.T) {
	tests := []struct {
		name    string
		weekday Weekday
		want    string
	}{{
		"sunday",
		0,
		"Sun",
	}, {
		"saturday",
		6,
		"Sat",
	}, {
		"next sunday",
		7,
		"Sun",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.weekday.String()
			Assertf(t, got == test.want, "got %q, expected %q", got, test.want)
		})
	}
}

func TestParseHours(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Hour
		wantErr error
	}{{
		"start of the day",
		"00:00",
		0,
		nil,
	}, {
		"almost end of the day",
		"23:59",
		1439,
		nil,
	}, {
		"end of the day",
		"24:00",
		1440,
		nil,
	}, {
		"invalid format",
		"invalid format",
		0,
		ErrorParseHour,
	}, {
		"input out of range",
		"24:01",
		0,
		ErrorParseHour,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseHour(test.input)
			Assertf(t, errors.Is(err, test.wantErr), "got %v, expected %v", err, test.wantErr)
			Assertf(t, got == test.want, "got %d, expected %d", got, test.want)
		})
	}
}

func TestHours_String(t *testing.T) {
	tests := []struct {
		name  string
		hours Hour
		want  string
	}{{
		"start of the day",
		0,
		"00:00",
	}, {
		"almost end of the day",
		1439,
		"23:59",
	}, {
		"end of the day",
		1440,
		"24:00",
	}, {
		"next day",
		1441,
		"00:01",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.hours.String()
			Assertf(t, got == test.want, "got %q, expected %q", got, test.want)
		})
	}
}

func TestParseBusinessHours(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *BusinessHours
		wantErr error
	}{{
		"multiple weekday business hours",
		"Mon-Fri 09:00-17:00",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		nil,
	}, {
		"single weekday business hours",
		"Mon 09:00-17:00",
		&BusinessHours{1, 1, 9 * 60, 17 * 60, time.UTC},
		nil,
	}, {
		"business hours with location",
		"Mon-Fri 09:00-17:00 Europe/Berlin",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, locEuropeBerlin},
		nil,
	}, {
		"invalid business hours format",
		"invalid format",
		nil,
		ErrorParseBusinessHours,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseBusinessHours(test.input)
			Assertf(t, errors.Is(err, test.wantErr), "got %q, expected %q", err, test.wantErr)
			Assertf(t, reflect.DeepEqual(got, test.want), "got %#v, expected %#v", got, test.want)
		})
	}
}

func TestBusinessHours_Contains(t *testing.T) {
	tests := []struct {
		name string
		bh   *BusinessHours
		time time.Time // INFO: the first weekday of the year 2006 was a sunday
		want bool
	}{{
		"inside weekdays 9to5",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-02 13:00"),
		true,
	}, {
		"day outside weekdays 9to5",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-01 13:00"),
		false,
	}, {
		"hour outside weekdays 9to5",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-02 05:00"),
		false,
	}, {
		"inside weekend 9to5",
		&BusinessHours{6, 7, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-01 13:00"),
		true,
	}, {
		"day outside weekend 9to5",
		&BusinessHours{6, 7, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-02 13:00"),
		false,
	}, {
		"hour outside weekend 9to5",
		&BusinessHours{6, 7, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-01 05:00"),
		false,
	}, {
		"inside start day of 5to1",
		&BusinessHours{1, 5, 17 * 60, 25 * 60, time.UTC},
		MustParseTime("2006-01-02 19:00"),
		true,
	}, {
		"inside end day of 5to1",
		&BusinessHours{1, 5, 17 * 60, 25 * 60, time.UTC},
		MustParseTime("2006-01-03 00:00"),
		true,
	}, {
		"hour outside start day of 5to1",
		&BusinessHours{1, 5, 17 * 60, 25 * 60, time.UTC},
		MustParseTime("2006-01-02 15:00"),
		false,
	}, {
		"hour outside end day of 5to1",
		&BusinessHours{1, 5, 17 * 60, 25 * 60, time.UTC},
		MustParseTime("2006-01-02 03:00"),
		false,
	}, {
		"at start of 9to5",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-02 09:00"),
		true,
	}, {
		"at end of 9to5",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		MustParseTime("2006-01-02 17:00"),
		false,
	}, {
		"at start of day after a complete day",
		&BusinessHours{1, 1, 0 * 60, 24 * 60, time.UTC},
		MustParseTime("2006-01-03 00:00"),
		false,
	}, {
		"end of day before a complete day",
		&BusinessHours{1, 1, 0 * 60, 24 * 60, time.UTC},
		MustParseTime("2006-01-02 00:00").Add(24 * time.Hour),
		false,
	}, {
		"hour inside 9to5 due to location",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, locUTCPlusTwo},
		MustParseTime("2006-01-02 08:00"),
		true,
	}, {
		"hour outside 9to5 due to location",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, locUTCPlusTwo},
		MustParseTime("2006-01-02 16:00"),
		false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.bh.ContainsTime(test.time)
			Assertf(t, got == test.want, "got: %t, want: %t", got, test.want)
		})
	}
}

func TestBusinessHours_String(t *testing.T) {
	tests := []struct {
		name string
		bh   *BusinessHours
		want string
	}{{
		"weekdays 9to5",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, nil},
		"Mon-Fri 09:00-17:00",
	}, {
		"weekdays 9to5 with location",
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
		"Mon-Fri 09:00-17:00 UTC",
	}, {
		"single weekday 9to5",
		&BusinessHours{1, 1, 9 * 60, 17 * 60, nil},
		"Mon 09:00-17:00",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.bh.String()
			Assertf(t, got == test.want, "got %q, expected %q", got, test.want)
		})
	}
}

func TestBusinessHours_MarshalJSON(t *testing.T) {
	got, err := json.Marshal(&struct {
		BH *BusinessHours `json:"bh"`
	}{
		&BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC},
	})
	want := `{"bh":"Mon-Fri 09:00-17:00 UTC"}`

	Assertf(t, errors.Is(err, nil), "got %v, expected nil", err)
	Assertf(t, string(got) == want, "got %q, expected %q", got, want)
}

func TestBusinessHours_UnmarshalJSON(t *testing.T) {
	var got struct {
		BH *BusinessHours `json:"bh"`
	}
	err := json.Unmarshal([]byte(`{"bh":"Mon-Fri 09:00-17:00"}`), &got)
	want := &BusinessHours{1, 5, 9 * 60, 17 * 60, time.UTC}

	Assertf(t, errors.Is(err, nil), "got %v, expected nil", err)
	Assertf(t, reflect.DeepEqual(got.BH, want), "got %#v, expected %#v", got.BH, want)
}

var (
	locEuropeBerlin, _ = time.LoadLocation("Europe/Berlin")
	locUTCPlusTwo      = time.FixedZone("UTC+02:00", 2*60*60)
)

// Assertf errors if the test clause fails with format and args
func Assertf(t *testing.T, clause bool, format string, args ...interface{}) {
	t.Helper()
	if !clause {
		t.Errorf(format, args...)
	}
}

// MustParseTime parses a time of the format "2006-01-02 15:04" or panics
func MustParseTime(in string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", in)
	if err != nil {
		panic(err)
	}
	return t
}
