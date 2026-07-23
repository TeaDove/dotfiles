package cli

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v3"
)

func CommandDePII(_ context.Context, _ *cli.Command) error {
	input, err := utils.ReadSTDINUntilEOF()
	if err != nil {
		return errors.Wrap(err, "read from stdin or pipe")
	}

	result := depiiMultiline(input)
	fmt.Print(result)
	return nil
}

func depiiMultiline(input string) string {
	uuidRegex := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	alfanum20Regex := regexp.MustCompile(`(?:[^a-z0-9]|^)[a-z0-9]{20}(?:[^a-z0-9]|$)`)

	uuidMap := make(map[string]int)
	alfanum20Map := make(map[string]int)

	uuidCounter := 0
	alfanum20Counter := 0

	result := input

	// Replace UUIDs
	result = uuidRegex.ReplaceAllStringFunc(result, func(uuid string) string {
		lowerUUID := strings.ToLower(uuid)
		if idx, exists := uuidMap[lowerUUID]; exists {
			return fmt.Sprintf("{UUID-%d}", idx)
		}
		idx := uuidCounter
		uuidMap[lowerUUID] = idx
		uuidCounter++
		return fmt.Sprintf("{UUID-%d}", idx)
	})

	// Replace alfanum20
	result = alfanum20Regex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract leading and trailing non-alphanumeric characters
		var prefix, suffix string
		id := match

		if len(id) > 0 && !isAlphaNum(rune(id[0])) {
			prefix = string(id[0])
			id = id[1:]
		}
		if len(id) > 0 && !isAlphaNum(rune(id[len(id)-1])) {
			suffix = string(id[len(id)-1])
			id = id[:len(id)-1]
		}

		if idx, exists := alfanum20Map[id]; exists {
			return fmt.Sprintf("%s{ALFANUM20-%d}%s", prefix, idx, suffix)
		}
		idx := alfanum20Counter
		alfanum20Map[id] = idx
		alfanum20Counter++
		return fmt.Sprintf("%s{ALFANUM20-%d}%s", prefix, idx, suffix)
	})

	return result
}

func depii(input string) string {
	uuidRegex := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	alfanum20Regex := regexp.MustCompile(`(?:[^a-z0-9]|^)[a-z0-9]{20}(?:[^a-z0-9]|$)`)

	uuidMap := make(map[string]int)
	alfanum20Map := make(map[string]int)

	uuidCounter := 0
	alfanum20Counter := 0

	result := input

	// Replace UUIDs
	result = uuidRegex.ReplaceAllStringFunc(result, func(uuid string) string {
		lowerUUID := strings.ToLower(uuid)
		if idx, exists := uuidMap[lowerUUID]; exists {
			return fmt.Sprintf("{UUID-%d}", idx)
		}
		idx := uuidCounter
		uuidMap[lowerUUID] = idx
		uuidCounter++
		return fmt.Sprintf("{UUID-%d}", idx)
	})

	// Replace alfanum20
	result = alfanum20Regex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract leading and trailing non-alphanumeric characters
		var prefix, suffix string
		id := match

		if len(id) > 0 && !isAlphaNum(rune(id[0])) {
			prefix = string(id[0])
			id = id[1:]
		}
		if len(id) > 0 && !isAlphaNum(rune(id[len(id)-1])) {
			suffix = string(id[len(id)-1])
			id = id[:len(id)-1]
		}

		if idx, exists := alfanum20Map[id]; exists {
			return fmt.Sprintf("%s{ALFANUM20-%d}%s", prefix, idx, suffix)
		}
		idx := alfanum20Counter
		alfanum20Map[id] = idx
		alfanum20Counter++
		return fmt.Sprintf("%s{ALFANUM20-%d}%s", prefix, idx, suffix)
	})

	return result
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
