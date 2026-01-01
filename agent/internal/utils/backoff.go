package utils

import (
    "math/rand"
    "time"
)

func ExponentialBackoff(attempt int, base, max time.Duration, jitter float64) time.Duration {
    if attempt < 0 {
        attempt = 0
    }
    d := base * (1 << attempt)
    if d > max {
        d = max
    }
    if jitter > 0 {
        j := rand.Float64()*2 - 1
        d = time.Duration(float64(d) * (1 + jitter*j))
        if d < 0 {
            d = base
        }
    }
    return d
}

