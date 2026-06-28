package main

import (
    "fmt"
    "os"

    "architecture-bricks/cmd/commands"
)

func main() {
    if err := commands.Execute(); err != nil {
        _, _ = fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
