package main

import (
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	keyQueryString   = "query-string"
	keySince         = "since"
	keyStart         = "start"
	keyEnd           = "end"
	keyDurationQuota = "duration-quota"
	keyVerbose       = "verbose"
)

func init() {
	RootCmd.AddCommand(&QueryCmd, &ListCmd, &BulkCmd, &VersionCmd)

	var (
		now   = time.Now()
		since = time.Minute * 5
	)
	composeopt(RootCmd.PersistentFlags(), []opt{
		{optname: "query-string", key: keyQueryString, defval: "", envname: "QUERY_STRING", desc: "query string"},
		{optname: "end", key: keyEnd, defval: iso8601utc(now), envname: "", desc: "end time"},
		{optname: "since", key: keySince, defval: since, envname: "", desc: "since"},
		{optname: "start", key: keyStart, defval: "", envname: "", desc: "start time"},
		{optname: "duration-quota", key: keyDurationQuota, defval: time.Hour * 24 * 1, envname: "", desc: "duration-quota"},
		{optname: "verbose", key: keyVerbose, defval: false, envname: "VERBOSE", desc: "verbosely"},
	})

	cobra.OnInitialize(func() {
		viper.SetConfigName(".cwinsights")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()
		viper.ReadInConfig()
		if viper.GetBool(keyVerbose) {
			logger.Printf = log.Printf
		}
	})
}

var RootCmd = cobra.Command{
	Use: "cwinsights",
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

func either(file *os.File) func(r io.Reader) io.Reader {
	return func(r io.Reader) io.Reader {
		if terminal.IsTerminal(int(file.Fd())) {
			return r
		}
		return file
	}
}

func parseTime(s string) time.Time {
	layouts := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"2006-01-02-07:00",
		"2006-01-02T00:00:00-07:00",
		"20060102",
		"20060102-07:00",
	}
	for _, e := range layouts {
		t, err := time.Parse(e, s)
		if err == nil {
			return t
		}
	}
	log.Fatalf("cannot parse '%v'", s)
	return time.Time{}
}

func startEndTime() (start, end time.Time) {
	var (
		since = viper.GetDuration(keySince)
		s     = viper.GetString(keyStart)
		e     = viper.GetString(keyEnd)
	)
	end = parseTime(e)
	if s == "" {
		start = end.Add(-since)
	} else {
		start = parseTime(s)
	}
	return
}

func iso8601utc(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func checkDurationQuota(d time.Duration) {
	if d <= 0 {
		log.Fatalf("the given duration is zero or negative. %v", d)
	}

	q := viper.GetDuration(keyDurationQuota)
	if d > q {
		log.Fatalf("the given duration exceeds the quota. %v > %v", d, q)
	}

}
