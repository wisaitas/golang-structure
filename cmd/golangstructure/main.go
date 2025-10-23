package main

import (
	"github.com/wisaitas/github.com/wisaitas/golang-structure/internal/golangstructure/initial"
)

func main() {
	app := initial.New()

	app.Run()

	app.Shutdown()
}
