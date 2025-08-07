package datetime

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/omniful/api-gateway/constants"
	"github.com/omniful/api-gateway/pkg/utils"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/nullable"
	"gopkg.in/guregu/null.v4"
)

const (
	DefaultTimeFormat             = "January 02, 2006 3:04 PM"
	DefaultTime                   = "January 01, 0001 12:00 AM"
	DefaultDate                   = "January 01, 0001"
	DateFormat                    = "02-01-2006"
	DateFormatInYYMMDD            = "2006-01-02"
	DateFormatInMMDDYY            = "January 02, 2006"
	DateFormatInHHMM              = "03:04 PM"
	DefaultDateTimeFormat         = "2006-01-02 15:04"
	DefaultDateTimeFormatDDMMYYYY = "02-01-2006 15:04"
)

func FormatTimeForKeys(ctx context.Context, m map[string]any, keys []string) {
	if m == nil {
		return
	}

	for _, key := range keys {
		timeStr, ok := m[key].(string)
		if !ok {
			continue
		}

		m[key] = FormatTime(ctx, timeStr)
	}
}

func FormatTime(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}
	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	// Convert the time to the user's timezone
	location := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	if err != nil {
		log.WithError(err).Error("cannot load user timezone location")
		return inputTime
	}

	timeLocal := timeUTC.In(location)
	var timeFormat string
	if len(format) == 1 {
		timeFormat = format[0]
	} else {
		timeFormat = DefaultTimeFormat
	}

	// Format the time as a string
	formattedTime = timeLocal.Format(timeFormat)
	if strings.Contains(formattedTime, DefaultDate) {
		return ""
	}
	return formattedTime
}

func FormatTimeWithoutTimeZone(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	// Parse the input time string into a time.Time value in UTC timezone
	timeLocal, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	var timeFormat string
	if len(format) == 1 {
		timeFormat = format[0]
	} else {
		timeFormat = DefaultTimeFormat
	}

	// Format the time as a string
	formattedTime = timeLocal.Format(timeFormat)
	if strings.Contains(formattedTime, DefaultDate) {
		return ""
	}
	return formattedTime
}

func GetTimeFromDate(ctx context.Context, date null.String) (respTime null.Int, err error) {
	if !date.Valid {
		return null.IntFromPtr(nil), nil
	}

	layout := DateFormat

	location := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	if err != nil {
		log.WithError(err).Error("cannot load user timezone location")
		location = time.UTC
	}

	if date.Valid {
		from, err := time.ParseInLocation(layout, date.String, location)
		if err != nil {
			return null.Int{}, err
		}
		respTime = null.IntFrom(from.Unix())
	}

	return respTime, nil
}

func GetTimeForDateRange(ctx context.Context, date null.String, dateType utils.DateType) (respTime null.String, err error) {
	var (
		location *time.Location
		ok       bool
	)
	if !date.Valid {
		return null.StringFromPtr(nil), nil
	}

	layout := DateFormat

	if location, ok = ctx.Value(constants.UserTimeZoneLocation).(*time.Location); !ok {
		log.WithError(err).Error("cannot load user timezone location")
		location = time.UTC
	}

	if dateType == utils.FromDate {
		from, err := time.ParseInLocation(layout, date.String, location)
		if err != nil {
			return null.String{}, err
		}
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, location)
		respTime = nullable.NewNullableString(strconv.FormatInt(from.Unix(), 10))
	}

	if dateType == utils.ToDate {
		to, err := time.ParseInLocation(layout, date.String, location)
		if err != nil {
			return null.String{}, err
		}
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, location)
		respTime = nullable.NewNullableString(strconv.FormatInt(to.Unix(), 10))
	}

	return
}

func GetTimeForDateRangeV2(ctx context.Context, dateTime null.String) (respTime null.String, err error) {
	var (
		location *time.Location
		ok       bool
	)
	if !dateTime.Valid {
		return null.StringFromPtr(nil), nil
	}

	// layout := DefaultDateTimeFormat
	layout := DateFormat + " 15:04"

	if location, ok = ctx.Value(constants.UserTimeZoneLocation).(*time.Location); !ok {
		log.WithError(err).Error("cannot load user timezone location")
		location = time.UTC
	}

	parsedDateTime, err := time.ParseInLocation(layout, dateTime.String, location)
	if err != nil {
		return null.String{}, err
	}
	parsedDateTime = time.Date(parsedDateTime.Year(), parsedDateTime.Month(), parsedDateTime.Day(), parsedDateTime.Hour(), parsedDateTime.Minute(), 0, 0, location)
	respTime = nullable.NewNullableString(strconv.FormatInt(parsedDateTime.Unix(), 10))

	return
}

func GetDateTimeForDateRange(ctx context.Context, dateTime null.String, dateType utils.DateType) (respTime null.String, err error) {
	if !dateTime.Valid {
		return null.StringFromPtr(nil), nil
	}

	layout := DateFormat + " 15:04" // Assuming DateFormat is "02-01-2006" and time format is 15:04

	location := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	if location == nil {
		log.Error("user timezone location is not set, defaulting to UTC")
		location = time.UTC
	}

	if dateType == utils.FromDate {
		from, err := time.ParseInLocation(layout, dateTime.String, location)
		if err != nil {
			return null.String{}, err
		}
		respTime = nullable.NewNullableString(strconv.FormatInt(from.Unix(), 10))
	}

	if dateType == utils.ToDate {
		to, err := time.ParseInLocation(layout, dateTime.String, location)
		if err != nil {
			return null.String{}, err
		}
		to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond) // Add 23 hours, 59 minutes, 59 seconds, and nanoseconds to get end of day
		respTime = nullable.NewNullableString(strconv.FormatInt(to.Unix(), 10))
	}

	return
}

func FormatDate(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}
	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	// Format the time as a string
	formattedTime = timeUTC.Format(DateFormat)
	return formattedTime
}

func FormatDateInYYMMDD(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}
	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	// Format the time as a string
	formattedTime = timeUTC.Format(DateFormatInYYMMDD)
	return formattedTime
}

func FormatDateInMMDDYY(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}
	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	// Format the time as a string
	formattedTime = timeUTC.Format(DateFormatInMMDDYY)
	return formattedTime
}

func FormatTimeToUTC(ctx context.Context, inputTimeStr string, format ...string) (formattedTime string) {
	var timeFormat string
	if len(format) == 1 {
		timeFormat = format[0]
	} else {
		timeFormat = DefaultDateTimeFormat
	}

	location := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	inputTime, err := time.ParseInLocation(timeFormat, inputTimeStr, location)
	if err != nil {
		return
	}

	utcTime := inputTime.UTC()
	formattedTime = utcTime.Format(DefaultDateTimeFormat)
	return
}

func FormatDateInHHMM(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}
	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	location := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	if err != nil {
		log.WithError(err).Error("cannot load user timezone location")
		return inputTime
	}

	timeLocal := timeUTC.In(location)

	// Format the time as a string
	formattedTime = timeLocal.Format(DateFormatInHHMM)
	return formattedTime
}

func FormatDateWithTimeZone(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}

	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	location, ok := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	if !ok {
		log.WithError(err).Error("cannot load user timezone location")
		return inputTime
	}

	// Convert parsed time to user's location
	localTime := timeUTC.In(location)

	// Format the time as a string
	formattedTime = localTime.Format(DateFormat)
	return formattedTime
}

func FormatDateTimeInDDMMYYYY(ctx context.Context, inputTime string, format ...string) (formattedTime string) {
	if inputTime == "" {
		return ""
	}

	// Parse the input time string into a time.Time value in UTC timezone
	timeUTC, err := time.Parse(time.RFC3339Nano, inputTime)
	if err != nil {
		return inputTime
	}

	location, ok := ctx.Value(constants.UserTimeZoneLocation).(*time.Location)
	if !ok {
		log.WithError(err).Error("cannot load user timezone location")
		return inputTime
	}

	// Convert parsed time to user's location
	localTime := timeUTC.In(location)

	// Format the time as a string
	formattedTime = localTime.Format(DefaultDateTimeFormatDDMMYYYY)
	return formattedTime
}
