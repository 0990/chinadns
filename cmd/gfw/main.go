package main

import "github.com/0990/chinadns/gfwlist"

const GFWLIST_URL = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"

func main() {
	gfwlist.CreateGFWListFile(GFWLIST_URL, "gfwlist.txt")
}
