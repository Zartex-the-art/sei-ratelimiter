package main

import (
	"fmt"
	"time"

	"github.com/Zartex-the-art/sei-ratelimiter/internal/limiter"
)

func main() {
	l := limiter.NewFixedWindowLimiter(3, 10*time.Second)

	for i := 1; i <= 5; i++ {
		allowed := l.Allow("hari")

		if allowed {
			fmt.Println("request", i, "allowed")
		} else {
			fmt.Println("request", i, "blocked")
		}
	}
}
