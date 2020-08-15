package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	keyQueryString   = "query-string"
	keyLogGroup      = "log-group"
	keyBefore        = "before"
	keyFull          = "full"
	keyFzf           = "fzf"
	keyDurationQuota = "duration-quota"
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigName(".cwinsight")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()
		viper.ReadInConfig()
	})

	RootCmd.AddCommand(&QueryCmd, &ListCmd, &BulkCmd)

	config(RootCmd.PersistentFlags(), []opt{
		{optname: "query-string", key: keyQueryString, defval: "", envname: "QUERY_STRING", desc: "query string"},
		{optname: "before", key: keyBefore, defval: time.Duration(0), envname: "BEFORE", desc: "before"},
		{optname: "duration-quota", key: keyDurationQuota, defval: time.Hour * 24 * 3, envname: "", desc: "duration-quota"},
	})
	config(QueryCmd.Flags(), []opt{
		{optname: "log-group", key: keyLogGroup, defval: "", envname: "LOG_GROUP", desc: "log group"},
		{optname: "fzf", key: keyFzf, defval: false, envname: "", desc: "fzf"},
	})
	config(ListCmd.Flags(), []opt{
		{optname: "full", key: keyFull, defval: false, envname: "", desc: "full"},
	})
}

var RootCmd = cobra.Command{
	Use: "cwinsight",
}

var cfg aws.Config

func main() {
	c, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg = c
	RootCmd.Execute()
}

func fzf(gs []cloudwatchlogs.LogGroup) (*cloudwatchlogs.LogGroup, error) {
	idx, err := fuzzyfinder.Find(
		gs,
		func(i int) string {
			return *gs[i].LogGroupName
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			retention := int64(0)
			if gs[i].RetentionInDays != nil {
				retention = *gs[i].RetentionInDays
			}
			var (
				name  = *gs[i].LogGroupName
				ctime = time.Unix(*gs[i].CreationTime/1000, 0)
				size  = *gs[i].StoredBytes
			)
			return fmt.Sprintf(`log-group        : %s
creation-time    : %v
stored-bytes     : %d
retention-in-days: %d days`, name, ctime, size, retention)
		}))
	if err != nil {
		return nil, err
	}
	return &gs[idx], nil
}

func listLogGroups() ([]cloudwatchlogs.LogGroup, error) {
	svc := cloudwatchlogs.New(cfg)
	res, err := svc.DescribeLogGroupsRequest(&cloudwatchlogs.DescribeLogGroupsInput{}).Send(context.Background())
	if err != nil {
		return nil, err
	}
	return res.LogGroups, nil
}

func either(file *os.File) func(r io.Reader) io.Reader {
	return func(r io.Reader) io.Reader {
		if terminal.IsTerminal(int(file.Fd())) {
			return r
		}
		return file
	}
}

func startEndTime() (start, end time.Time) {
	var (
		now    = time.Now()
		before = viper.GetDuration(keyBefore)
	)
	//log.Printf("%v %v", before, now.Add(-before))
	start = now.Add(-before)
	end = now
	return
}
