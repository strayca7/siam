package main

import (
	"os"

	"github.com/strayca7/siam/internal/apiserver"
	"github.com/strayca7/siam/pkg/app"
	namev1 "github.com/strayca7/siam/staging/src/api/name/v1"
)

func main() {
	application := apiserver.NewApp(namev1.APIServer)
	code := app.Run(application)
	os.Exit(code)
}
