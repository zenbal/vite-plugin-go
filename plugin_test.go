package viteplugin

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// test using the example manifest from https://vitejs.dev/guide/backend-integration
const TestManifest = `
{
  "_shared-CPdiUi_T.js": {
    "file": "assets/shared-ChJ_j-JJ.css",
    "src": "_shared-CPdiUi_T.js"
  },
  "_shared-B7PI925R.js": {
    "file": "assets/shared-B7PI925R.js",
    "name": "shared",
    "css": ["assets/shared-ChJ_j-JJ.css"]
  },
  "baz.js": {
    "file": "assets/baz-B2H3sXNv.js",
    "name": "baz",
    "src": "baz.js",
    "isDynamicEntry": true
  },
  "views/bar.js": {
    "file": "assets/bar-gkvgaI9m.js",
    "name": "bar",
    "src": "views/bar.js",
    "isEntry": true,
    "imports": ["_shared-B7PI925R.js"],
    "dynamicImports": ["baz.js"]
  },
  "views/foo.js": {
    "file": "assets/foo-BRBmoGS9.js",
    "name": "foo",
    "src": "views/foo.js",
    "isEntry": true,
    "imports": ["_shared-B7PI925R.js"],
    "css": ["assets/foo-5UjPuW-k.css"]
  }
}
`

func TestGenerate(t *testing.T) {
	var m Manifest
	if err := json.Unmarshal([]byte(TestManifest), &m); err != nil {
		t.Error("failed to unmarshal manifest")
	}

	eps := m.GetEntryPoints()
	if len(eps) != 2 {
		t.Errorf("expected GetEntryPoints to return map with two entries but got map of length %d", len(eps))
	}

	keys := make([]string, 0, len(eps))
	for k := range eps {
		keys = append(keys, k)
	}
	expectedKeys := []string{"views/bar.js", "views/foo.js"}

	sort.Strings(keys)
	sort.Strings(expectedKeys)

	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("expected entrypoint keys to equal '%v' but got '%v'", expectedKeys, keys)
	}

	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name: "Bar entry point",
			key:  "views/bar.js",
			expected: `
				<link rel="stylesheet" href="assets/shared-ChJ_j-JJ.css"/>
				<script type="module" src="assets/bar-gkvgaI9m.js"></script>
				<link rel="modulepreload" href="assets/shared-B7PI925R.js"/>
			`,
		},
		{
			name: "Foo entry point",
			key:  "views/foo.js",
			expected: `
				<link rel="stylesheet" href="assets/foo-5UjPuW-k.css"/>
				<link rel="stylesheet" href="assets/shared-ChJ_j-JJ.css"/>
				<script type="module" src="assets/foo-BRBmoGS9.js"></script>
				<link rel="modulepreload" href="assets/shared-B7PI925R.js"/>
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			if err := generate(&b, &m, eps[tc.key]); err != nil {
				t.Fatalf("generate failed: %v", err)
			}

			result := b.String()

			if removeWhitespace(result) != removeWhitespace(tc.expected) {
				t.Errorf("Expected generate to return:\n%v\nfor entrypoint '%s' but got:\n%v", removeWhitespace(tc.expected), tc.key, removeWhitespace(result))
			}
		})
	}
}

func TestAddPrefix(t *testing.T) {
	var m Manifest
	if err := json.Unmarshal([]byte(TestManifest), &m); err != nil {
		t.Error("failed to unmarshal manifest")
	}
	prefix := "/static/"
	m.AddPrefix(prefix)

	for _, chunk := range m.GetEntryPoints() {
		if chunk.File != "" && chunk.File[:len(prefix)] != prefix {
			t.Errorf("Expected Chunk '%v's File field to start with '%v' but got '%v'", chunk.Name, prefix, chunk.File)
		}
		for _, css := range chunk.Css {
			if css[:len(prefix)] != prefix {
				t.Errorf("Expected Css entry from Chunk '%v' to start with '%v' but got '%v'", chunk.Name, prefix, css)
			}
		}
	}
}

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			return -1
		}
		return r
	}, s)
}
