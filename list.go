package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
