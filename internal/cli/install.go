package cli

import (
	"context"
	"dotfiles/internal/cli/utils/closeutils"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

var dotfilesDirs = [...]string{"./dotfiles-configs"}

var mergeConfigs = mapset.NewSet[string]()

var codexMappings = [...][2]string{
	{".claude/settings.json", ""},
	{".claude/hooks/*", ""},
	{".claude/CLAUDE.md", ".codex/AGENTS.md"},
	{".claude/skills/*/SKILL.md", ".codex/prompts/*.md"},
	{".claude/*", ".codex/*"},
}

func CommandInstall(_ context.Context, _ *cli.Command) error { //nolint: gocognit // FIXME
	var dofilesPath string

	for _, dir := range dotfilesDirs {
		_, err := os.Stat(dir)
		if !errors.Is(err, fs.ErrNotExist) {
			dofilesPath = dir

			break
		}
	}

	if dofilesPath == "" {
		return errors.Errorf("no dotfiles-configs dir found")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user home dir")
	}

	err = filepath.Walk(dofilesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "walk")
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dofilesPath, path)
		if err != nil {
			return errors.Wrap(err, "rel path")
		}

		dst := filepath.Join(homeDir, rel)

		err = os.MkdirAll(filepath.Dir(dst), 0o755)
		if err != nil {
			return errors.Wrap(err, "mkdir target dir")
		}

		if mergeConfigs.Contains(rel) {
			err = mergeJSONFile(path, dst)
			if err != nil {
				return errors.Wrapf(err, "merge %s", rel)
			}

			return nil
		}

		err = overwriteFile(path, dst, info.Mode())
		if err != nil {
			return errors.Wrapf(err, "copy %s", rel)
		}

		codexRel, ok := codexPath(rel)
		if ok {
			err = installCodexFile(path, filepath.Join(homeDir, codexRel), info.Mode())
			if err != nil {
				return errors.Wrapf(err, "install codex %s", rel)
			}
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "install files")
	}

	color.Green("Dotfiles installed from %s to %s", dofilesPath, homeDir)

	return nil
}

func codexPath(rel string) (string, bool) {
	for _, mapping := range codexMappings {
		pattern, target := mapping[0], mapping[1]

		prefix, suffix, hasStar := strings.Cut(pattern, "*")
		if !hasStar {
			if rel != pattern {
				continue
			}

			return target, target != ""
		}

		if len(rel) < len(prefix)+len(suffix) ||
			!strings.HasPrefix(rel, prefix) ||
			!strings.HasSuffix(rel, suffix) {
			continue
		}

		if target == "" {
			return "", false
		}

		captured := rel[len(prefix) : len(rel)-len(suffix)]

		return strings.Replace(target, "*", captured, 1), true
	}

	return "", false
}

func installCodexFile(src, dst string, mode fs.FileMode) error {
	err := os.MkdirAll(filepath.Dir(dst), 0o755)
	if err != nil {
		return errors.Wrap(err, "mkdir target dir")
	}

	err = overwriteFile(src, dst, mode)
	if err != nil {
		return errors.Wrap(err, "copy")
	}

	return nil
}

func overwriteFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open src")
	}
	defer closeutils.MustClose(in)

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return errors.Wrap(err, "open dst")
	}

	_, err = io.Copy(out, in)
	if err != nil {
		_ = out.Close()

		return errors.Wrap(err, "copy contents")
	}

	err = out.Close()
	if err != nil {
		return errors.Wrap(err, "close dst")
	}

	return nil
}

func mergeJSONFile(src, dst string) error {
	srcData, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, "read fragment")
	}

	var srcMap map[string]any

	err = json.Unmarshal(srcData, &srcMap)
	if err != nil {
		return errors.Wrap(err, "parse fragment")
	}

	dstMap := map[string]any{}

	dstData, err := os.ReadFile(dst)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errors.Wrap(err, "read target")
	}

	if err == nil {
		err = json.Unmarshal(dstData, &dstMap)
		if err != nil {
			return errors.Wrap(err, "parse target")
		}
	}

	merged := deepMerge(dstMap, srcMap)

	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal merged")
	}

	err = os.WriteFile(dst, append(out, '\n'), 0o644) //nolint: gosec // FIXME
	if err != nil {
		return errors.Wrap(err, "write target")
	}

	return nil
}

func deepMerge(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}

	for key, srcVal := range src {
		dstVal, ok := dst[key]
		if ok {
			dstObj, dstIsObj := dstVal.(map[string]any)
			srcObj, srcIsObj := srcVal.(map[string]any)

			if dstIsObj && srcIsObj {
				dst[key] = deepMerge(dstObj, srcObj)

				continue
			}
		}

		dst[key] = srcVal
	}

	return dst
}
