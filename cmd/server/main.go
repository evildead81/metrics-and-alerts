package main

import (
	"github.com/evildead81/metrics-and-alerts/internal/server/instance"
)

func main() {
	instance.NewServerInstance("8080").Run()
}
