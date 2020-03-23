package main

import "strings"

func nsInstance(url string) string {
	n := strings.TrimLeft(url, "https://")
	n = strings.TrimLeft(n, "http://")
	n = strings.Trim(n, " /")
	shortname := strings.Split(n, `.`)
	if len(shortname) > 0 {
		return shortname[0]
	}
	return n
}
