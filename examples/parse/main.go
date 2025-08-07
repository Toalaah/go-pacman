package main

import (
	_ "embed"
	"fmt"
	"log"

	pacman "github.com/toalaah/go-pacman"
)

func main() {
	pkg, err := pacman.QueryPackage("ffmpeg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n\n", pkg)
	b, err := pkg.MarshalText()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", string(b))
}
