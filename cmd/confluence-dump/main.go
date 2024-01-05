/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package main

import "log"

func main() {
	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}
