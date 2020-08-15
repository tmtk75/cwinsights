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

func composeopt(fs *pflag.FlagSet, opts []opt) {
	for _, e := range opts {
		if e.key == "" {
			log.Fatalf(`empty key. %v`, e)
		}
		if e.optname == "" {
			log.Fatalf(`empty optname. %v`, e)
		}

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
		if e.envname != "" {
			viper.BindEnv(e.key, e.envname)
		}
	}
}
