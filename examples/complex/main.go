package main

import (
	"context"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

func main() {
	fdk.Run(context.Background(), newHandler)
}
