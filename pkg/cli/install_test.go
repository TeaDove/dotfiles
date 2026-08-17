package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepMerge(t *testing.T) {
	t.Run("fragment adds keys and preserves local ones", func(t *testing.T) {
		dst := map[string]any{"model": "opus", "theme": "dark"}
		src := map[string]any{"hooks": map[string]any{"Stop": []any{"x"}}}

		got := deepMerge(dst, src)

		assert.Equal(t, "opus", got["model"])
		assert.Equal(t, "dark", got["theme"])
		assert.Equal(t, map[string]any{"Stop": []any{"x"}}, got["hooks"])
	})

	t.Run("src wins on shared scalar key", func(t *testing.T) {
		dst := map[string]any{"model": "old"}
		src := map[string]any{"model": "new"}

		got := deepMerge(dst, src)

		assert.Equal(t, "new", got["model"])
	})

	t.Run("nested objects merge key-by-key", func(t *testing.T) {
		dst := map[string]any{"hooks": map[string]any{"Stop": []any{"s"}}}
		src := map[string]any{"hooks": map[string]any{"Notification": []any{"n"}}}

		got := deepMerge(dst, src)

		hooks := got["hooks"].(map[string]any)
		assert.Equal(t, []any{"s"}, hooks["Stop"])
		assert.Equal(t, []any{"n"}, hooks["Notification"])
	})

	t.Run("arrays are replaced, not appended", func(t *testing.T) {
		dst := map[string]any{"list": []any{"a", "b"}}
		src := map[string]any{"list": []any{"c"}}

		got := deepMerge(dst, src)

		assert.Equal(t, []any{"c"}, got["list"])
	})

	t.Run("dst-only key survives", func(t *testing.T) {
		dst := map[string]any{"keep": 1.0}
		src := map[string]any{"add": 2.0}

		got := deepMerge(dst, src)

		assert.Equal(t, 1.0, got["keep"])
		assert.Equal(t, 2.0, got["add"])
	})
}

func TestMergeJSONFile(t *testing.T) {
	fragment := `{"hooks":{"Stop":[{"cmd":"notify"}]}}`

	writeFile := func(t *testing.T, dir, name, content string) string {
		t.Helper()
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))

		return p
	}

	readJSON := func(t *testing.T, path string) map[string]any {
		t.Helper()
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(data, &m))

		return m
	}

	t.Run("existing target keeps local keys and gains fragment", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "fragment.json", fragment)
		dst := writeFile(t, dir, "settings.json", `{"model":"opus","theme":"dark"}`)

		require.NoError(t, mergeJSONFile(src, dst))

		got := readJSON(t, dst)
		assert.Equal(t, "opus", got["model"])
		assert.Equal(t, "dark", got["theme"])
		assert.Contains(t, got, "hooks")
	})

	t.Run("re-running is idempotent", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "fragment.json", fragment)
		dst := writeFile(t, dir, "settings.json", `{"model":"opus"}`)

		require.NoError(t, mergeJSONFile(src, dst))
		first, err := os.ReadFile(dst)
		require.NoError(t, err)

		require.NoError(t, mergeJSONFile(src, dst))
		second, err := os.ReadFile(dst)
		require.NoError(t, err)

		assert.Equal(t, first, second)
	})

	t.Run("missing target writes fragment verbatim", func(t *testing.T) {
		dir := t.TempDir()
		src := writeFile(t, dir, "fragment.json", fragment)
		dst := filepath.Join(dir, "settings.json")

		require.NoError(t, mergeJSONFile(src, dst))

		got := readJSON(t, dst)
		assert.Contains(t, got, "hooks")
		assert.NotContains(t, got, "model")
	})
}
