package config

import (
	"os"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Config represents config
type Config struct {
	StartTime           *time.Time
	EndTime             *time.Time
	LogGroupNameFilter  *regexp.Regexp
	LogStreamNameFilter *regexp.Regexp
	LogStreamNamePrefix string
	FilterPattern       string
	Profile             string
	Code                string
	Region              string
	Timestamps          bool
	EventID             bool
	NoLogGroupName      bool
	NoLogStreamName     bool
	MaxLength           int
	Color               bool
}

// New returns Config
func New(c *cli.Context) (*Config, error) {
	logGroupName := c.Args().First()
	if logGroupName == "" {
		cli.ShowAppHelp(c)
		os.Exit(1)
	}
	logStreamName := c.String("stream")

	now := time.Now()

	since := c.String("since")
	startDuration, err := time.ParseDuration(since)
	var startTime time.Time
	if err != nil {
		startTime, err = time.Parse(time.RFC3339, since)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid 'since' time format")
		}
	} else {
		startTime = now.Add(-startDuration)
	}

	if startTime.Unix() < 0 {
		return nil, errors.New("'since' must be specified to be after January 1, 1970 UTC")
	}

	var endTime *time.Time
	end := c.String("end")
	if end != "" {
		endDuration, err := time.ParseDuration(end)
		if err != nil {
			e, err := time.Parse(time.RFC3339, end)
			if err != nil {
				return nil, errors.Wrap(err, "Invalid 'end' time format")
			}
			endTime = &e
		} else {
			e := now.Add(-endDuration)
			endTime = &e
		}
		if endTime.Unix() < 0 {
			return nil, errors.New("'end' must be specified to be after January 1, 1970 UTC")
		}
	}

	config := &Config{
		StartTime:           &startTime,
		EndTime:             endTime,
		LogGroupNameFilter:  regexp.MustCompile(logGroupName),
		LogStreamNameFilter: regexp.MustCompile(logStreamName),
		LogStreamNamePrefix: c.String("stream-prefix"),
		FilterPattern:       c.String("filter"),
		Profile:             c.String("profile"),
		Code:                c.String("code"),
		Region:              c.String("region"),
		Timestamps:          c.Bool("timestamps"),
		EventID:             c.Bool("event-id"),
		NoLogGroupName:      c.Bool("no-log-group"),
		NoLogStreamName:     c.Bool("no-log-stream"),
		MaxLength:           c.Int("max-length"),
		Color:               c.Bool("color"),
	}

	return config, nil
}
