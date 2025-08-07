package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	pacman "github.com/toalaah/go-pacman"
)

func main() {
	q := os.Args[1]
	fmt.Printf("Query: %s\n", q)
	pkg, err := pacman.QueryPackage(q)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found Package: %+v\n\n", pkg)
	b, err := pkg.MarshalText()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", string(b))
}
