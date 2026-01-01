package main

import (
    "fmt"
    "os"
)

import (
    "pulse/agent/internal/config"
    "pulse/agent/internal/core"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Configuration error")
        os.Exit(1)
    }
    app := core.NewApp(cfg)
    if err := app.Prepare(); err != nil {
        fmt.Fprintln(os.Stderr, "Startup error")
        os.Exit(1)
    }
    if err := app.Run(); err != nil {
        os.Exit(2)
    }
}

