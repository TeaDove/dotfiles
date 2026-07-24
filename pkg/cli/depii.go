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

	result := depii(input)
	fmt.Print(result)
	return nil
}

func depii(input string) string {
	input = replaceUUID(input)
	input = replaceAlfaNum20(input)
	input = replaceLongDomains(input)

	return input
}

func replaceAlfaNum20(input string) string {
	counter := make(map[string]int)
	total := 0
	reg := regexp.MustCompile(`(?:[^a-z0-9]|^)[a-z0-9]{20}(?:[^a-z0-9]|$)`)
	return reg.ReplaceAllStringFunc(input, func(match string) string {
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

		if idx, exists := counter[id]; exists {
			return fmt.Sprintf("%s{ALFANUM20-%d}%s", prefix, idx, suffix)
		}
		idx := total
		counter[id] = idx
		total++
		return fmt.Sprintf("%s{ALFANUM20-%d}%s", prefix, idx, suffix)
	})
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func replaceUUID(input string) string {
	counter := make(map[string]int)
	total := 0
	reg := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	return reg.ReplaceAllStringFunc(input, func(uuid string) string {
		lowerUUID := strings.ToLower(uuid)
		if idx, exists := counter[lowerUUID]; exists {
			return fmt.Sprintf("{UUID-%d}", idx)
		}
		idx := total
		counter[lowerUUID] = idx
		total++
		return fmt.Sprintf("{UUID-%d}", idx)
	})
}

func replaceLongDomains(input string) string {
	counter := make(map[string]int)
	total := 0
	reg := regexp.MustCompile(`[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?){2,}`)

	return reg.ReplaceAllStringFunc(input, func(match string) string {
		lowerDomain := strings.ToLower(match)
		if idx, exists := counter[lowerDomain]; exists {
			return fmt.Sprintf("{DOMAIN-%d}", idx)
		}
		idx := total
		counter[lowerDomain] = idx
		total++
		return fmt.Sprintf("{DOMAIN-%d}", idx)
	})
}
