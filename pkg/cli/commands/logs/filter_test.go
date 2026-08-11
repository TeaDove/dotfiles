package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter_NoPatterns_MatchesEverything(t *testing.T) {
	t.Parallel()

	var f Filter

	assert.True(t, f.Matches("anything at all"))
	assert.Equal(t, "", f.IncludePattern())
	assert.Equal(t, "", f.ExcludePattern())
}

func TestFilter_IncludeOnly(t *testing.T) {
	t.Parallel()

	var f Filter

	err := f.SetInclude("ERROR")
	require.NoError(t, err)

	assert.True(t, f.Matches("this is an ERROR"))
	assert.False(t, f.Matches("this is fine"))
	assert.Equal(t, "ERROR", f.IncludePattern())
}

func TestFilter_ExcludeOnly(t *testing.T) {
	t.Parallel()

	var f Filter

	err := f.SetExclude("DEBUG")
	require.NoError(t, err)

	assert.False(t, f.Matches("this is DEBUG noise"))
	assert.True(t, f.Matches("this is fine"))
	assert.Equal(t, "DEBUG", f.ExcludePattern())
}

func TestFilter_IncludeAndExclude_ExcludeWins(t *testing.T) {
	t.Parallel()

	var f Filter

	err := f.SetInclude("ERROR")
	require.NoError(t, err)

	err = f.SetExclude("DEBUG")
	require.NoError(t, err)

	assert.False(t, f.Matches("ERROR: DEBUG stuff"))
	assert.True(t, f.Matches("ERROR: real problem"))
	assert.False(t, f.Matches("no match here"))
}

func TestFilter_ClearingPattern(t *testing.T) {
	t.Parallel()

	var f Filter

	err := f.SetInclude("ERROR")
	require.NoError(t, err)

	err = f.SetInclude("")
	require.NoError(t, err)

	assert.True(t, f.Matches("no longer filtered"))
	assert.Equal(t, "", f.IncludePattern())
}

func TestFilter_CaseInsensitive(t *testing.T) {
	t.Parallel()

	var f Filter

	err := f.SetInclude("error")
	require.NoError(t, err)

	assert.True(t, f.Matches("ERROR: something broke"))
	assert.True(t, f.Matches("error: something broke"))
	assert.True(t, f.Matches("ErRoR: mixed case"))
}

func TestFilter_InvalidPattern_LeavesPreviousUnchanged(t *testing.T) {
	t.Parallel()

	var f Filter

	err := f.SetInclude("ERROR")
	require.NoError(t, err)

	err = f.SetInclude("(unclosed")
	require.Error(t, err)

	assert.Equal(t, "ERROR", f.IncludePattern())
	assert.True(t, f.Matches("still an ERROR"))
}
