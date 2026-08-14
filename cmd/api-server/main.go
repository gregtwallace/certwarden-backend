package main

import (
	"certwarden-backend/pkg/domain/app"
	"os"
)

func main() {
	os.Exit(app.Run())
}
