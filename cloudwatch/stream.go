package cloudwatch

import (
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
)

// LogStream is log stream
type LogStream struct {
	Name               *string
	LastEventTimestamp *int64
}

// ListStreams lists stream names matching the specified filter
func (cwl *Client) ListStreams(groupName string, since int64) (streams []*LogStream, err error) {
	// If LogStreamNamePrefix is specified, log streams can not be sorted by LastEventTimestamp.
	// https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatchlogs/#DescribeLogStreamsInput
	if cwl.config.LogStreamNamePrefix != "" {
		since = 0
	}
	streams = []*LogStream{}
	fn := func(res *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
		hasUpdatedStream := false
		minLastIngestionTime := int64(math.MaxInt64)
		for _, stream := range res.LogStreams {
			if *stream.LastIngestionTime < minLastIngestionTime {
				minLastIngestionTime = *stream.LastIngestionTime
			}
			if !cwl.config.LogStreamNameFilter.MatchString(*stream.LogStreamName) {
				continue
			}
			// Use LastIngestionTime because LastEventTimestamp is updated slowly...
			if stream.LastIngestionTime == nil || *stream.LastIngestionTime < since {
				continue
			}
			hasUpdatedStream = true
			streams = append(streams, &LogStream{
				Name:               stream.LogStreamName,
				LastEventTimestamp: stream.LastIngestionTime,
			})
		}
		if cwl.config.LogStreamNamePrefix != "" {
			return true
		}
		if minLastIngestionTime >= since {
			return true
		}
		return hasUpdatedStream
	}

	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(groupName)}
	if cwl.config.LogStreamNamePrefix != "" {
		input.LogStreamNamePrefix = aws.String(cwl.config.LogStreamNamePrefix)
	} else {
		input.OrderBy = aws.String("LastEventTime")
		input.Descending = aws.Bool(true)
	}

	err = cwl.client.DescribeLogStreamsPages(input, fn)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return streams, nil
			} else if awsErr.Code() == "ThrottlingException" {
				time.Sleep(500 * time.Millisecond)
				return nil, nil
			}
		}
		return nil, errors.Wrap(err, "Failed to DescribeLogStreams")
	}
	return streams, nil
}
