package main

import (
	"log"

	"github.com/gonejack/thunderbird-rss-html/cmd"
)

func main() {
	var c cmd.Converter
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
