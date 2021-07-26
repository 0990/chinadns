package main

import (
	"fmt"
	"log"
	"net/url"
	"testing"
)

func TestUrl(t *testing.T) {
	u, err := url.Parse("d.zhuanfou.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u.Hostname())
}
