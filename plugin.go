package viteplugin

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

type Plugin struct {
	FileSystem  fs.FS // outDir
	Manifest    Manifest
	EntryPoints map[string]*Chunk
}

func New(fileSystem fs.FS, manifestPath string) (*Plugin, error) {
	plugin := &Plugin{
		FileSystem: fileSystem,
	}
	if err := plugin.LoadManifest(manifestPath); err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}
	plugin.EntryPoints = plugin.Manifest.GetEntryPoints()
	return plugin, nil
}

type Manifest map[string]*Chunk

type Chunk struct {
	File           string   `json:"file"`
	Name           string   `json:"name"`
	Src            string   `json:"src"`
	IsEntry        bool     `json:"isEntry"`
	IsDynamicEntry bool     `json:"isDynamicEntry"`
	Imports        []string `json:"imports"`
	DynamicImports []string `json:"dynamicImports"`
	Css            []string `json:"css"`
}

func (m *Manifest) GetEntryPoint() (*Chunk, error) {
	for _, chunk := range *m {
		if chunk.IsEntry {
			return chunk, nil
		}
	}
	return nil, fmt.Errorf("entry chunk not found")
}

func (m *Manifest) GetEntryPoints() map[string]*Chunk {
	entryPoints := make(map[string]*Chunk)
	for key, chunk := range *m {
		if chunk.IsEntry {
			entryPoints[key] = chunk
		}
	}
	return entryPoints
}

func (plugin *Plugin) LoadManifest(path string) error {
	mFile, err := plugin.FileSystem.Open(path)
	if err != nil {
		return err
	}
	var m Manifest
	if err := json.NewDecoder(mFile).Decode(&m); err != nil {
		return err
	}
	entryPoints := m.GetEntryPoints()
	if len(entryPoints) == 0 {
		return fmt.Errorf("can't load manifest with no entrypoint")
	}
	plugin.Manifest = m
	return nil
}

func (plugin *Plugin) RawHTML(entry *Chunk) (string, error) {
	var b strings.Builder
	if err := generate(&b, &plugin.Manifest, entry); err != nil {
		return "", fmt.Errorf("failed to generate HTML: %w", err)
	}
	return b.String(), nil
}

func generate(b *strings.Builder, manifest *Manifest, entry *Chunk) error {
	if err := genCss(b, manifest, entry); err != nil {
		return err
	}
	writeTag(b, "script", map[string]string{
		"type": "module",
		"src":  entry.File,
	})
	b.WriteString("\n")
	if err := genPreload(b, manifest, entry); err != nil {
		return err
	}
	return nil
}

func genCss(b *strings.Builder, manifest *Manifest, curr *Chunk) error {
	for _, path := range curr.Css {
		writeTag(b, "link", map[string]string{
			"rel":  "stylesheet",
			"href": path,
		})
		b.WriteString("\n")
	}
	for _, importKey := range curr.Imports {
		i, ok := (*manifest)[importKey]
		if !ok {
			return fmt.Errorf("import '%s' not found in manifest", importKey)
		}
		genCss(b, manifest, i)
	}
	return nil
}

func genPreload(b *strings.Builder, manifest *Manifest, curr *Chunk) error {
	for _, importKey := range curr.Imports {
		i, ok := (*manifest)[importKey]
		if !ok {
			return fmt.Errorf("import '%s' not found in manifest", importKey)
		}
		writeTag(b, "link", map[string]string{
			"rel":  "modulepreload",
			"href": i.File,
		})
		b.WriteString("\n")
		genPreload(b, manifest, i)
	}
	return nil
}

func writeTag(b *strings.Builder, tag string, attributes map[string]string) {
	attributeOrder := []string{"type", "src", "rel", "href"}
	b.WriteString("<")
	b.WriteString(tag)
	b.WriteString(" ")
	for i, attr := range attributeOrder {
		if value, ok := attributes[attr]; ok {
			writeAttr(b, attr, value)
			if i < len(attributes)-1 {
				b.WriteString(" ")
			}
		}
	}
	if tag == "script" {
		b.WriteString("></script>")
	} else {
		b.WriteString("/>")
	}
}

func writeAttr(b *strings.Builder, attr, value string) {
	b.WriteString(attr)
	b.WriteString("=\"")
	b.WriteString(value)
	b.WriteString("\"")
}
