package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	keyLogGroup = "log-group"
	keyFzf      = "fzf"
)

func init() {
	composeopt(QueryCmd.Flags(), []opt{
		{optname: "log-group", key: keyLogGroup, defval: "", envname: "LOG_GROUP", desc: "log group"},
		{optname: "fzf", key: keyFzf, defval: false, envname: "", desc: "fzf"},
	})
}

var QueryCmd = cobra.Command{
	Use: "query",
	Run: func(c *cobra.Command, args []string) {
		gn := viper.GetString(keyLogGroup)
		if viper.GetBool(keyFzf) {
			gs, err := listLogGroups()
			if err != nil {
				log.Fatal(err)
			}
			lg, err := fzf(gs)
			gn = *lg.LogGroupName
		}

		start, end := startEndTime()
		r := Query(viper.GetString(keyQueryString), gn, start, end)
		fmt.Printf("%v", r)
	},
}

func Query(qs, group string, start, end time.Time) *cloudwatchlogs.GetQueryResultsResponse {
	//log.Printf("qs: %v, group: %v, start: %v, end: %v", qs, group, start, end)
	svc := cloudwatchlogs.New(cfg)
	res, err := svc.StartQueryRequest(&cloudwatchlogs.StartQueryInput{
		LogGroupName: aws.String(group),
		QueryString:  aws.String(qs),
		StartTime:    aws.Int64(start.Unix()),
		EndTime:      aws.Int64(end.Unix()),
	}).Send(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}

	//fmt.Printf("%v\n", res)

wait:
	r, err := svc.GetQueryResultsRequest(&cloudwatchlogs.GetQueryResultsInput{
		QueryId: res.QueryId,
	}).Send(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}
	if r.Status != "Complete" {
		//fmt.Printf("%v\n", r)
		time.Sleep(time.Second * 1)
		goto wait
	}

	return r
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
