package businesshours

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrorParseBusinessHours is returned when the input string couldn't be parsed to valid business hours.
	ErrorParseBusinessHours = errors.New("couldn't parse businesshours")
	// ErrorParseHour is returned when the input string couldn't be parsed to a valid hour.
	ErrorParseHour = errors.New("couldn't parse hour")
	// ErrorParseWeekday is returned when the input string couldn't be parsed to a valid weekday.
	ErrorParseWeekday = errors.New("couldn't parse weekday")
)

// Weekday represents a day of the week (Sun = 0, ...).
type Weekday int

// Hour represent minutes within a 1440 minute day (0 = 00:00, ..., 1440 = 24:00).
type Hour int

// BusinessHours describes weekly business hours for a specific location.
type BusinessHours struct {
	startDay, endDay Weekday
	startHr, endHr   Hour
	loc              *time.Location
}

var (
	daysOfWeek = map[string]int{
		"Sun": 0,
		"Mon": 1,
		"Tue": 2,
		"Wed": 3,
		"Thu": 4,
		"Fri": 5,
		"Sat": 6,
	}
	daysOfWeekInv = map[int]string{
		0: "Sun",
		1: "Mon",
		2: "Tue",
		3: "Wed",
		4: "Thu",
		5: "Fri",
		6: "Sat",
	}
)

// ParseWeekday converts a 3 letter string for a weekday ("Sun", ..., "Sat") into its weekday index number.
func ParseWeekday(in string) (Weekday, error) {
	out, ok := daysOfWeek[in]
	if !ok {
		return 0, fmt.Errorf("%w: %s is not a valid weekday", ErrorParseWeekday, in)
	}
	return Weekday(out), nil
}

// String implements the fmt.Stringer interface.
func (d Weekday) String() string {
	return daysOfWeekInv[int(d)%7]
}

// validHourRE is used for validating the hour format "HH:MM" ("00:00", ..., "23:59", "24:00")
var validHourRE *regexp.Regexp = regexp.MustCompile("^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)")

// ParseHour converts a hour string of the format "HH:MM" into the number of minutes elapsed in the day. Valid hour
// formats are "00:00", ..., "23:59", "24:00"
func ParseHour(in string) (Hour, error) {
	if !validHourRE.MatchString(in) {
		return 0, fmt.Errorf("%w: %q invalid format", ErrorParseHour, in)
	}

	components := strings.Split(in, ":")
	if len(components) != 2 {
		return 0, fmt.Errorf("%w: %q invalid format", ErrorParseHour, in)
	}

	hours, err := strconv.Atoi(components[0])
	if err != nil {
		return 0, fmt.Errorf("%w: converting hour %q", ErrorParseHour, in)
	}

	minutes, err := strconv.Atoi(components[1])
	if err != nil {
		return 0, fmt.Errorf("%w: converting minute %q", ErrorParseHour, in)
	}

	if hours < 0 || hours > 24 || minutes < 0 || minutes > 60 {
		return 0, fmt.Errorf("%w: hour %q out of range", ErrorParseHour, in)
	}

	// Hour are stored as minutes elapsed in the day, so multiply hour by 60.
	return Hour(hours*60 + minutes), nil
}

// String implements the fmt.Stringer interface.
func (h Hour) String() string {
	var hours int
	if h != 1440 {
		hours = int(h % 1440 / 60)
	} else {
		hours = 24
	}

	minutes := h % 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

// ParseBusinessHours converts a business hour string of a format lke "Mon-Fri 09:00-17:00 Europe/Berlin" into
// BusinessHours. When a single weekday is provided the start and end day will be the same. In case the location
// is omitted UTC is assumed.
func ParseBusinessHours(in string) (*BusinessHours, error) {
	components := strings.Split(in, " ")
	if !(len(components) == 2 || len(components) == 3) {
		return nil, fmt.Errorf("%w: invalid format %q", ErrorParseBusinessHours, in)
	}

	startDay, endDay, err := parseWeekdays(components[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorParseBusinessHours, err)
	}

	startHr, endHr, err := parseHours(components[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorParseBusinessHours, err)
	}

	var loc *time.Location
	if len(components) == 3 {
		loc, err = time.LoadLocation(components[2])
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrorParseBusinessHours, err)
		}
	} else {
		loc = time.UTC
	}

	return &BusinessHours{
		startDay, endDay,
		startHr, endHr,
		loc,
	}, nil
}

// ContainsTime checks if a given time.Time is inside the business hours.
func (bh *BusinessHours) ContainsTime(t time.Time) bool {
	tin := t.In(bh.loc)
	tinDay := Weekday(tin.Weekday())
	tinHr := Hour(tin.Hour()*60 + tin.Minute())
	return bh.containsWeekday(tinDay) &&
		bh.containsHour(tinHr)
}

// containsWeekday checks if a given Weekday is inside the business hours.
func (bh *BusinessHours) containsWeekday(day Weekday) bool {
	endDay := bh.endDay
	// increment endDay by a day, when bh spans 2days
	if bh.endHr > 1440 {
		endDay = endDay + 1
	}

	// increment day by a week, when bh spans 2weeks and day might be on the next week
	if !(bh.startDay <= day || bh.endDay <= 6) {
		day = day + 7
	}

	return bh.startDay <= day && day <= endDay
}

// containsHour checks if a given Hour is inside the business hours.
func (bh *BusinessHours) containsHour(hr Hour) bool {
	// incrementHr by a day, when bh spans 2days and hr might be on the next day
	if !(bh.startHr <= hr || bh.endHr <= 1440) {
		hr = hr + 1440
	}

	return bh.startHr <= hr && hr < bh.endHr
}

// String implements the fmt.Stringer interface.
func (bh *BusinessHours) String() string {
	weekdays := fmt.Sprintf("%s-%s", bh.startDay, bh.endDay)
	if bh.startDay == bh.endDay {
		weekdays = bh.startDay.String()
	}

	location := ""
	if bh.loc != nil {
		location = fmt.Sprintf(" %s", bh.loc)
	}

	return fmt.Sprintf("%s %s-%s%s", weekdays, bh.startHr, bh.endHr, location)
}

// MarshalJSON implements the json.Marshaler interface.
func (bh *BusinessHours) MarshalJSON() ([]byte, error) {
	return json.Marshal(bh.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (bh *BusinessHours) UnmarshalJSON(in []byte) error {
	var str string
	if err := json.Unmarshal(in, &str); err != nil {
		return err
	}

	businesshours, err := ParseBusinessHours(str)
	if err != nil {
		return err
	}
	*bh = *businesshours
	return nil
}

// parseWeekdays parses start and end weekday of the weekday range format "Day-Day". If only a single day is provided
// start and end weekday are the same.
func parseWeekdays(in string) (Weekday, Weekday, error) {
	var start, end Weekday
	var err error

	days := strings.Split(in, "-")
	if !(len(days) == 1 || len(days) == 2) {
		return 0, 0, fmt.Errorf("invalid weekday range format %q", in)
	}

	if start, err = ParseWeekday(days[0]); err != nil {
		return 0, 0, err
	}

	if len(days) == 1 {
		return start, start, nil
	}

	if end, err = ParseWeekday(days[1]); err != nil {
		return 0, 0, err
	}

	return start, end, nil
}

// parseHours parses the start and end hour of the hour range format "HH:MM-HH:MM"
func parseHours(in string) (Hour, Hour, error) {
	var start, end Hour
	var err error

	hours := strings.Split(in, "-")
	if !(len(hours) == 2) {
		return 0, 0, fmt.Errorf("invalid hour range format %q", in)
	}

	if start, err = ParseHour(hours[0]); err != nil {
		return 0, 0, err
	}

	if end, err = ParseHour(hours[1]); err != nil {
		return 0, 0, err
	}

	return start, end, nil
}
