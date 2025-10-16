package git

import (
	"context"
	"dotfiles/pkg/cli/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type commitNameParams struct {
	fancyHour          string
	hostname           string
	commitChangedFiles int
	commitInsertions   int
	commitDeletions    int
}

func calcCommitMsg(params *commitNameParams) string {
	var msg strings.Builder
	msg.WriteString(params.fancyHour)
	msg.WriteString(" ")
	msg.WriteString(getCommitChangeFilesWord(params.commitChangedFiles))
	msg.WriteString(" ")
	msg.WriteString(getCommitDivWord(params.commitInsertions, params.commitDeletions))
	if params.hostname != "" {
		msg.WriteString(" from ")
		msg.WriteString(params.hostname)
	}

	return msg.String()
}

func getCommitMsg(ctx context.Context) string {
	var params commitNameParams

	params.fancyHour = hourToFancyName(time.Now().Hour())
	params.hostname = getHostname(ctx)
	params.commitChangedFiles, params.commitInsertions, params.commitDeletions = getCommitChanges(ctx)

	return calcCommitMsg(&params)
}

func getCommitDivWord(commitInsertions, commitDeletions int) string {
	div := float64(commitInsertions) / float64(commitDeletions)
	if div > 2 {
		return "edit"
	}
	if div > 0.5 {
		return "fix"
	}

	return "erase"
}

func getCommitChangeFilesWord(commitChangedFiles int) string {
	if commitChangedFiles < 3 {
		return "minor"
	}
	if commitChangedFiles < 15 {
		return "small"
	}
	if commitChangedFiles < 30 {
		return "major"
	}

	return "refactoring"
}
func getHostname(ctx context.Context) string {
	out, err := utils.ExecCommand(ctx, "hostname")
	if err != nil {
		return ""
	}

	return out
}

func hourToFancyName(hour int) string {
	if hour < 1 {
		return "midnight"
	}
	if hour < 5 {
		return "nightly"
	}
	if hour < 10 {
		return "early"
	}
	if hour < 18 {
		return "daily"
	}
	if hour < 23 {
		return "evening"
	}
	return "midnight"
}

func extractRegexpInt(pattern, s string) int {
	submatch := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if len(submatch) < 2 {
		return 0
	}

	v, err := strconv.Atoi(submatch[1])
	if err != nil {
		return 0
	}

	return v
}

func getCommitChanges(ctx context.Context) (int, int, int) {
	out, err := utils.ExecCommand(ctx, "git", "diff", "--staged", "--shortstat")
	if err != nil {
		return 0, 0, 0
	}

	return extractCommitChanges(out)
}

func extractCommitChanges(diffLine string) (int, int, int) {
	return extractRegexpInt(`(\d+) file`, diffLine),
		extractRegexpInt(`(\d+) insertion`, diffLine),
		extractRegexpInt(`(\d+) deletion`, diffLine)
}
