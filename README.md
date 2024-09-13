# vite-plugin-go

[Vite Backend Integration](https://vitejs.dev/guide/backend-integration) in a single file for Go web applications.

## Installation

Use the following command:

```bash
go get github.com/zenbal/vite-plugin-go
```

or just copy the contents of `plugin.go` into your project.

## Usage

Make sure that vite generates a `manifest.json`

```go
package main

import (
    "os"
    "github.com/zenbal/vite-plugin-go"
)

func main() {
    // instantiate plugin. manifestPath is relative to dist
    plugin, _ := viteplugin.New(os.DirFS("path/to/dist"), ".vite/manifest.json")
    // iterate over entrypoints
    for _, chunk := range plugin.EntryPoints {
        html, err := plugin.RawHTML(chunk)
        fmt.Println(html)
    }
    // or access specific ones
    html, err := plugin.RawHTML(plugin.EntryPoints["main.ts"])
}
```
