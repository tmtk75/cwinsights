package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	keyFull = "full"
)

func init() {
	composeopt(ListCmd.Flags(), []opt{
		{optname: "full", key: keyFull, defval: false, envname: "", desc: "full"},
	})
}

var ListCmd = cobra.Command{
	Use: "list",
	Run: func(c *cobra.Command, args []string) {
		List()
	},
}

func List() {
	gs, err := listLogGroups()
	if err != nil {
		log.Fatalf("%v", err)
	}

	if viper.GetBool(keyFull) {
		b, _ := json.Marshal(gs)
		fmt.Printf("%v", string(b))
		return
	}

	for _, e := range gs {
		fmt.Printf("%v\n", *e.LogGroupName)
	}
}

func listLogGroups() ([]cloudwatchlogs.LogGroup, error) {
	svc := cloudwatchlogs.New(cfg)
	res, err := svc.DescribeLogGroupsRequest(&cloudwatchlogs.DescribeLogGroupsInput{}).Send(context.Background())
	if err != nil {
		return nil, err
	}
	return res.LogGroups, nil
}
