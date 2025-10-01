package assert

import (
	"log/slog"
	"os"
)

func Assert(truth bool, msg string){
	if !truth {
		slog.Error(msg)
		os.Exit(1)
	}
}

