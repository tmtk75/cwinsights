package main

import (
	"log"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type opt struct {
	optname string
	key     string
	defval  interface{}
	envname string
	desc    string
}

func config(fs *pflag.FlagSet, opts []opt) {
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
