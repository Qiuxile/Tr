package main

import (
	"os"

	"Tr/internal/tr"
)

func main() {
	tr.Run(os.Args, assetsFS)
}
