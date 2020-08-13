package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	keyQueryString = "query-string"
	keyLogGroup    = "log-group"
	keyBefore      = "before"
	keyFull        = "full"
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetConfigName(".cwinsight")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()
		viper.ReadInConfig()
	})

	RootCmd.AddCommand(&QueryCmd, &ListCmd)

	type opt struct {
		optname string
		key     string
		defval  interface{}
		envname string
		desc    string
	}

	config := func(fs *pflag.FlagSet, opts []opt) {
		for _, e := range opts {
			switch t := e.defval.(type) {
			case string:
				fs.String(e.optname, t, e.desc)
			case bool:
				fs.Bool(e.optname, t, e.desc)
			case time.Duration:
				fs.Duration(e.optname, t, e.desc)
			default:
				log.Fatalf(`unsupported type. "%v" %v`, e.optname, e.defval)
			}
			viper.BindPFlag(e.key, fs.Lookup(e.optname))
			viper.BindEnv(e.key, e.envname)
		}
	}

	config(QueryCmd.Flags(), []opt{
		{optname: "query-string", key: keyQueryString, defval: "", envname: "QUERY_STRING", desc: "query string"},
		{optname: "log-group", key: keyLogGroup, defval: "", envname: "LOG_GROUP", desc: "log group"},
		{optname: "before", key: keyBefore, defval: time.Duration(0), envname: "BEFORE", desc: "before"},
	})
	config(ListCmd.Flags(), []opt{
		{optname: "full", key: keyFull, defval: false, envname: "", desc: "full"},
	})
}

var RootCmd = cobra.Command{
	Use: "cwinsight",
}

var QueryCmd = cobra.Command{
	Use: "query",
	Run: func(c *cobra.Command, args []string) {
		Query(viper.GetString(keyQueryString), viper.GetString(keyLogGroup))
	},
}

var ListCmd = cobra.Command{
	Use: "list",
	Run: func(c *cobra.Command, args []string) {
		List()
	},
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

func Query(qs, group string) {
	svc := cloudwatchlogs.New(cfg)
	now := time.Now()
	before := viper.GetDuration(keyBefore)
	res, err := svc.StartQueryRequest(&cloudwatchlogs.StartQueryInput{
		LogGroupName: aws.String(group),
		QueryString:  aws.String(qs),
		StartTime:    aws.Int64(now.Truncate(before).Unix()),
		EndTime:      aws.Int64(now.Unix()),
	}).Send(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("%v\n", res)

wait:
	r, err := svc.GetQueryResultsRequest(&cloudwatchlogs.GetQueryResultsInput{
		QueryId: res.QueryId,
	}).Send(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}
	if r.Status != "Complete" {
		fmt.Printf("%v\n", r)
		time.Sleep(time.Second * 1)
		goto wait
	}

	fmt.Printf("%v\n", r)
}

func List() {
	svc := cloudwatchlogs.New(cfg)
	res, err := svc.DescribeLogGroupsRequest(&cloudwatchlogs.DescribeLogGroupsInput{}).Send(context.Background())
	if err != nil {
		log.Fatalf("%v", err)
	}

	if viper.GetBool(keyFull) {
		fmt.Printf("%v\n", res)
		return
	}

	for _, e := range res.LogGroups {
		fmt.Printf("%v\n", *e.LogGroupName)
	}
}
