package main

import (
	"github.com/peytoncasper/waypoint-plugin-gcs/registry"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
)

func main() {
	sdk.Main(sdk.WithComponents(
		&registry.Registry{},
	))
}
