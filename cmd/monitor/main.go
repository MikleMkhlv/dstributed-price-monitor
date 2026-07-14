package main

import (
	"bookmarks/dstributed-price-monitor/config"
	"fmt"
)

func main() {
	cfg := config.MustLoadConfig("conf_test.yaml")

	fmt.Println(cfg)
}
