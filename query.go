package main

import (
	"context"
	"encoding/json"
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
	keyGroupName = "group-name"
	keyFzf       = "fzf"
	keyRaw       = "raw"
)

func init() {
	composeopt(QueryCmd.Flags(), []opt{
		{optname: "group-name", key: keyGroupName, defval: "", envname: "GROUP_NAME", desc: "group name"},
		{optname: "fzf", key: keyFzf, defval: false, envname: "", desc: "fuzzyfinder"},
		{optname: "raw", key: keyRaw, defval: false, envname: "", desc: "print in raw format of AWS SDK"},
	})
}

var QueryCmd = cobra.Command{
	Use:   "query",
	Short: "execute query",
	Run: func(c *cobra.Command, args []string) {
		gn := viper.GetString(keyGroupName)
		if viper.GetBool(keyFzf) {
			gs, err := listLogGroups()
			if err != nil {
				log.Fatal(err)
			}
			lg, err := fzf(gs)
			if err != nil {
				log.Fatal(err)
			}
			gn = *lg.LogGroupName
		}

		start, end := startEndTime()
		r := Query(viper.GetString(keyQueryString), gn, start, end)
		f := format(r)
		b, err := json.Marshal(f)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", string(b))
	},
}

type QueryResult struct {
	*cloudwatchlogs.GetQueryResultsResponse
	QueryId string
}

func Query(qs, group string, start, end time.Time) *QueryResult {
	logger.Printf("group-name: %v", group)
	logger.Printf("query-string: %v", qs)
	logger.Printf("start: %v", iso8601utc(start))
	logger.Printf("end: %v", iso8601utc(end))
	logger.Printf("duration: %v", end.Sub(start))

	checkDurationQuota(end.Sub(start))

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

wait:
	r, err := svc.GetQueryResultsRequest(&cloudwatchlogs.GetQueryResultsInput{
		QueryId: res.QueryId,
	}).Send(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}
	if r.Status != "Complete" {
		time.Sleep(time.Second * 1)
		goto wait
	}

	return &QueryResult{
		GetQueryResultsResponse: r,
		QueryId:                 *res.QueryId,
	}
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

func format(r *QueryResult) interface{} {
	if viper.GetBool(keyRaw) {
		return r

	}

	a := make([]map[string]string, 0)
	for _, e := range r.Results {
		m := make(map[string]string)
		for _, v := range e {
			m[*v.Field] = *v.Value
		}
		a = append(a, m)
	}
	return struct {
		Results    []map[string]string
		Statistics interface{}
		QueryId    string
	}{
		Results:    a,
		Statistics: r.Statistics,
		QueryId:    r.QueryId,
	}
}
