package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var confFile = flag.String("c", "chinadns_err.json", "create from error")
var msg = flag.String("msg", "use china dns,but reply is abroad", "find key")
var output = flag.String("o", "gfwuser.txt.tmp", "output filename")

func main() {
	flag.Parse()

	f, err := os.Open(*confFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var line []byte
	var m map[string]string
	var domains map[string]struct{}
	for {
		data, prefix, err := r.ReadLine()
		if err != nil {
			break
		}
		if prefix {
			line = append(line, data...)
			continue
		}
		err = json.Unmarshal(data, &m)
		if err != nil {
			panic(err)
			continue
		}
		if m["msg"] != *msg {
			continue
		}

		domain := m["domain"]
		if domain != "" {
			domains[domain] = struct{}{}
		}

		line = line[:0]
	}

	of, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer of.Close()

	for domain, v := range domains {
		fmt.Println(v)
		of.WriteString(domain)
		of.Write([]byte{'\n'})
	}
}
