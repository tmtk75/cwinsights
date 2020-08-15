package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
