package main

import "strings"

func nsInstance(url string) string {
	n := strings.TrimLeft(url, "https://")
	n = strings.TrimLeft(n, "http://")
	n = strings.Trim(n, " /")
	return n
}
