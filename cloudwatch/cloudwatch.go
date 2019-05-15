package cloudwatch

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/briandowns/spinner"
	c "github.com/fatih/color"
	"github.com/knqyf263/utern/cache"
	"github.com/knqyf263/utern/color"
	"github.com/knqyf263/utern/config"
	"github.com/pkg/errors"
)

var seenGroup = new(sync.Map)
var seenStream = new(sync.Map)

// Client represents CloudWatch Logs client
type Client struct {
	client *cloudwatchlogs.CloudWatchLogs
	config *config.Config
}

type logEvent struct {
	logGroupName string
	event        *cloudwatchlogs.FilteredLogEvent
}

// NewClient creates a new instance of the CLoudWatch Logs client
func NewClient(conf *config.Config) *Client {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(conf.Region),
		},
	}
	sess := session.Must(session.NewSessionWithOptions(opts))
	return &Client{
		client: cloudwatchlogs.New(sess),
		config: conf,
	}
}

// Tail tails log
func (cwl *Client) Tail(ctx context.Context) error {
	start := make(chan struct{}, 1)
	ch := make(chan *logEvent, 1000)
	errch := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				cwl.print(event)
			}
		}
	}()

	go func() {
		apiTicker := time.NewTicker(250 * time.Millisecond)
		for range apiTicker.C {
			start <- struct{}{}
		}
	}()

	logGroupNames, err := cwl.ListGroups()
	if err != nil {
		return errors.Wrap(err, "Failed to list log groups")
	}

	for _, logGroupName := range logGroupNames {
		cwl.showNewGroup(logGroupName)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-start:
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Start()
		s.Suffix = " Fetching log streams..."

		streams, err := cwl.ListStreams(logGroupName, cwl.config.StartTime.Unix()*1000)
		if err != nil {
			return errors.Wrap(err, "Initial check failed")
		}
		s.Stop()

		for _, stream := range streams {
			cwl.showNewStream(logGroupName, stream)
		}
	}

	wg := &sync.WaitGroup{}
	for _, logGroupName := range logGroupNames {
		wg.Add(1)
		go func(groupName string) {
			defer wg.Done()
			if err := cwl.tail(ctx, groupName, start, ch, errch); err != nil {
				errch <- err
			}
		}(logGroupName)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		for {
			if len(ch) == 0 {
				close(done)
				break
			}
		}
	}()

	select {
	case <-ctx.Done():
		close(ch)
		return ctx.Err()
	case err := <-errch:
		return err
	case <-done:
	}
	return nil
}

func (cwl *Client) tail(ctx context.Context, logGroupName string,
	start chan struct{}, ch chan *logEvent, errch chan error) error {
	lastSeenTime := aws.Int64(cwl.config.StartTime.UTC().Unix() * 1000)

	fn := func(res *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
		for _, event := range res.Events {
			if cache.Cache.Load(logGroupName, event.EventId) {
				continue
			}
			cache.Cache.Store(logGroupName, event.EventId, event.IngestionTime)
			ch <- &logEvent{
				logGroupName: logGroupName,
				event:        event,
			}
		}

		if lastPage && len(res.Events) > 0 {
			lastSeenTime = res.Events[len(res.Events)-1].IngestionTime
			cache.Cache.Expire(logGroupName, lastSeenTime)
		}

		return true
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-start:
		}

		streams, err := cwl.ListStreams(logGroupName, *lastSeenTime)
		if err != nil {
			return err
		}

		streamNames := []*string{}
		for _, stream := range streams {
			streamNames = append(streamNames, stream.Name)
			cwl.showNewStream(logGroupName, stream)
		}

		if len(streamNames) == 0 {
			continue
		}

		if len(streamNames) > 100 {
			return errors.New("log streams exceed 100, so please filter log streams with '--stream' or '--stream-prefix' option")
		}

		input := &cloudwatchlogs.FilterLogEventsInput{
			LogGroupName:   aws.String(logGroupName),
			LogStreamNames: streamNames,
			Interleaved:    aws.Bool(true),
			StartTime:      lastSeenTime,
		}
		if cwl.config.FilterPattern != "" {
			input.FilterPattern = aws.String(cwl.config.FilterPattern)
		}
		if cwl.config.EndTime != nil {
			input.EndTime = aws.Int64(cwl.config.EndTime.UTC().Unix() * 1000)
		}

		err = cwl.client.FilterLogEventsPages(input, fn)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ThrottlingException" {
					log.Printf("Rate exceeded for %s. Wait for 500ms then retry.\n", logGroupName)
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}
			return errors.Wrap(err, "Unknown error while FilterLogEventsPages")
		}

		if cwl.config.EndTime != nil {
			return nil
		}
	}
}

func (cwl *Client) showNewGroup(groupName string) {
	if _, ok := seenGroup.Load(groupName); ok {
		return
	}
	g := color.GroupColor.Get(groupName).SprintFunc()
	p := c.New(c.FgHiGreen, c.Bold).SprintFunc()
	fmt.Fprintf(os.Stderr, "%s %s\n", p("+"), g(groupName))
	seenGroup.Store(groupName, struct{}{})
}

func (cwl *Client) showNewStream(groupName string, stream *LogStream) {
	groupStream := groupName + "+++" + *stream.Name
	if _, ok := seenStream.Load(groupStream); ok {
		return
	}
	g := color.GroupColor.Get(groupName).SprintFunc()
	s := color.StreamColor.Get(*stream.Name).SprintFunc()
	p := c.New(c.FgHiGreen, c.Bold).SprintFunc()
	t := formatUnixTime(*stream.LastEventTimestamp)
	fmt.Fprintf(os.Stderr, "%s %s › %s (%s)\n", p("+"), g(groupName), s(*stream.Name), t)
	seenStream.Store(groupStream, struct{}{})
}

func (cwl *Client) print(event *logEvent) {
	g := color.GroupColor.Get(event.logGroupName).SprintFunc()
	s := color.StreamColor.Get(*event.event.LogStreamName).SprintFunc()
	messages := []string{}
	if !cwl.config.NoLogGroupName {
		messages = append(messages, g(event.logGroupName))
	}
	if !cwl.config.NoLogStreamName {
		messages = append(messages, s(*event.event.LogStreamName))
	}
	if cwl.config.EventID {
		messages = append(messages, *event.event.EventId)
	}
	if cwl.config.Timestamps {
		t := formatUnixTime(*event.event.IngestionTime)
		messages = append(messages, t)
	}
	message := *event.event.Message
	if 0 < cwl.config.MaxLength && cwl.config.MaxLength < len(message) {
		message = message[:cwl.config.MaxLength]
	}
	messages = append(messages, message)
	fmt.Printf("%s\n", strings.Join(messages, " "))
}

func formatUnixTime(unixtime int64) string {
	sec := unixtime / 1000
	msec := unixtime % 1000
	t := time.Unix(sec, msec*1000)
	t = t.In(time.Local)
	return t.Format(time.RFC3339)
}
