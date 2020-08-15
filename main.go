package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

var BulkCmd = cobra.Command{
	Use:  "bulk [file]",
	Args: cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r := either(f)(os.Stdin)
		Bulk(viper.GetString(keyQueryString), r)
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

func Bulk(qs string, r io.Reader) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	l := strings.Split(strings.Trim(string(b), " \t\n"), "\n")
	s, e := startEndTime()
	d := e.Sub(s)
	log.Printf("d: %v", d)
	if d*time.Duration(len(l)) > viper.GetDuration(keyDurationQuota) {
		log.Fatalf("exceeded 24h, %v", d)
	}

	type Result struct {
		Response  *cloudwatchlogs.GetQueryResultsResponse
		GroupName string
	}
	res := make(chan *Result)
	var wg sync.WaitGroup
	f := func(lg string) {
		res <- &Result{Response: Query(qs, lg, s, e), GroupName: lg}
		wg.Done()
	}

	for _, e := range l {
		wg.Add(1)
		go f(e)
	}

	a := make([]*Result, 0)
	go func() {
		for e := range res {
			//fmt.Printf("%v\n", e)
			a = append(a, e)
		}
	}()
	wg.Wait()
	close(res)

	type Output struct {
		StartTime time.Time
		EndTime   time.Time
		Results   []*Result
	}
	bb, err := json.Marshal(Output{Results: a, StartTime: s, EndTime: e})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", string(bb))
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
