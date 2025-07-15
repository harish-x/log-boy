package pkg

import (
	"errors"
	"fmt"
	"time"
)

func ParseFormettedTimeString(t string) (string, error) {
	if t == "" {
		return "", fmt.Errorf("argument t is empty")
	}

	formattedTime, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return "", fmt.Errorf("invalid time format data")
	}
	return formattedTime.Format(time.RFC3339), nil

}

func FormateFilterDateIfExists(fromStr string, toStr string) (string, string, error) {
	var fromFormatted, toFormatted string
	if fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return "", "", errors.New("invalid FROM date format")
		}
		fromFormatted = from.Format(time.RFC3339)
	}

	if toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return "", "", errors.New("invalid TO date format")
		}
		toFormatted = to.Format(time.RFC3339)
	}

	return fromFormatted, toFormatted, nil
}

func ConvertStringToDate(d string) (time.Time, error) {
	layout := "02-01-2006-15"
	t, err := time.ParseInLocation(layout, d, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format")
	}
	return t, nil
}

func ConvertStringToEpochMillis(d string) (int64, error) {
	layout := "2006-01-02-15"
	t, err := time.ParseInLocation(layout, d, time.UTC)
	if err != nil {
		return 0, fmt.Errorf("invalid date format")
	}

	epochMillis := t.UnixNano() / int64(time.Millisecond)
	return epochMillis, nil
}

func ConvertEpochMillisToString(epochMillis int64) string {
	t := time.Unix(0, epochMillis*int64(time.Millisecond))
	return t.Format("2006-01-02-15")
}
