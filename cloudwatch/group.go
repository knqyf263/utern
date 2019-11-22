package cloudwatch

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
)

// ListGroups lists group names matching the specified filter
func (cwl *Client) ListGroups() (groupNames []string, err error) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = os.Stderr
	s.Start()
	s.Suffix = " Fetching log groups..."
	defer s.Stop()

	groupNames = []string{}
	fn := func(res *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		for _, group := range res.LogGroups {
			if !cwl.config.LogGroupNameFilter.MatchString(*group.LogGroupName) {
				continue
			}
			groupNames = append(groupNames, *group.LogGroupName)
		}
		return !lastPage
	}

	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	err = cwl.client.DescribeLogGroupsPages(input, fn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to DescribeLogGroups")
	}

	return groupNames, nil
}
