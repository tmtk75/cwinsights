package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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