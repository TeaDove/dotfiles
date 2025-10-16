package git

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExtractCommitChanges(t *testing.T) {
	changed, insertions, deleteions := extractCommitChanges("2 files changed, 14 insertions(+), 10 deletions(-)")
	assert.Equal(t, 2, changed)
	assert.Equal(t, 14, insertions)
	assert.Equal(t, 10, deleteions)

	changed, insertions, deleteions = extractCommitChanges("2 files changed, 1 deletions(-)")
	assert.Equal(t, 2, changed)
	assert.Equal(t, 0, insertions)
	assert.Equal(t, 1, deleteions)
}

func TestHourToFancyName(t *testing.T) {
	assert.Equal(t, "evening", hourToFancyName(time.Date(2025, 10, 16, 20, 0, 0, 0, time.UTC).Hour()))
}

func TestCalcCommitMsg(t *testing.T) {
	msg := calcCommitMsg(&commitNameParams{
		fancyHour:          "evening",
		hostname:           "linux",
		commitChangedFiles: 2,
		commitInsertions:   15,
		commitDeletions:    5,
	})
	assert.Equal(t, "evening minor edit from linux\n", msg)
}
