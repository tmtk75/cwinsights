package main

import (
	"io"
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
	keyBefore        = "before"
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
