package cli

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExtractRandomPartHexV7(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		u    uuid.UUID
		want string
	}{
		{
			name: "zero rand parts",
			u:    uuidFromBytes(0, 0, 0, 0, 0, 0, 0x70, 0, 0x80, 0, 0, 0, 0, 0, 0, 0),
			want: "0000000000000000000",
		},
		{
			name: "rand_b only",
			u:    uuidFromBytes(0, 0, 0, 0, 0, 0, 0x70, 0, 0x80, 0, 0, 0, 0, 0, 0, 1),
			want: "0000000000000000001",
		},
		{
			name: "rand_a only",
			u:    uuidFromBytes(0, 0, 0, 0, 0, 0, 0x71, 0, 0x80, 0, 0, 0, 0, 0, 0, 0),
			want: "0400000000000000000",
		},
		{
			name: "rand_a and rand_b",
			u:    uuidFromBytes(0, 0, 0, 0, 0, 0, 0x7a, 0xbc, 0xbf, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde),
			want: "2af3f123456789abcde",
		},
		{
			name: "max rand_a",
			u:    uuidFromBytes(0, 0, 0, 0, 0, 0, 0x7f, 0xff, 0x80, 0, 0, 0, 0, 0, 0, 0),
			want: "3ffc000000000000000",
		},
		{
			name: "max rand_b",
			u:    uuidFromBytes(0, 0, 0, 0, 0, 0, 0x70, 0, 0xbf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff),
			want: "0003fffffffffffffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractRandomPartHexV7(tt.u))
		})
	}
}

func TestNewV7FromTime(t *testing.T) {
	t.Parallel()

	t.Run("second precision", func(t *testing.T) {
		t.Parallel()

		input := time.Date(2025, 5, 5, 3, 3, 3, 0, time.UTC)
		u := newV7FromTime(input)

		assert.Equal(t, uuid.Version(7), u.Version())
		assert.Equal(t, uuid.RFC4122, u.Variant())

		sec, nsec := u.Time().UnixTime()
		got := time.Unix(sec, nsec).UTC()
		assert.Equal(t, input, got)

		milli := input.UnixMilli()
		assert.Equal(t, []byte{
			byte(milli >> 40),
			byte(milli >> 32),
			byte(milli >> 24),
			byte(milli >> 16),
			byte(milli >> 8),
			byte(milli),
			0x70,
			0,
		}, u[:8])
	})

	t.Run("sub-millisecond seq", func(t *testing.T) {
		t.Parallel()

		input := time.Date(2025, 5, 5, 3, 3, 3, 123456789, time.UTC)
		u := newV7FromTime(input)

		milli := input.UnixMilli()
		seq := (input.UnixNano() - milli*1_000_000) >> 8
		randA := (uint16(u[6]&0x0F) << 8) | uint16(u[7])

		assert.Equal(t, uint16(seq), randA)
		assert.Equal(t, byte(0x70)|byte(seq>>8)&0x0F, u[6])
		assert.Equal(t, byte(seq), u[7])
	})

	t.Run("time bytes stable for same input", func(t *testing.T) {
		t.Parallel()

		input := time.Date(2026, 1, 15, 12, 30, 45, 987654321, time.UTC)
		u1 := newV7FromTime(input)
		u2 := newV7FromTime(input)

		assert.Equal(t, u1[:8], u2[:8])
	})

	t.Run("random tail differs", func(t *testing.T) {
		t.Parallel()

		input := time.Date(2026, 1, 15, 12, 30, 45, 0, time.UTC)
		u1 := newV7FromTime(input)
		u2 := newV7FromTime(input)

		assert.NotEqual(t, u1[8:], u2[8:])
	})
}

func TestVerboseUUID(t *testing.T) {
	t.Parallel()

	input := time.Date(2025, 5, 5, 3, 3, 3, 0, time.UTC)
	u := newV7FromTime(input)

	out := verboseUUID(u)

	assert.Contains(t, out, "input:")
	assert.Contains(t, out, u.String())
	assert.Contains(t, out, "time:")
	assert.Contains(t, out, "time utc:")
	assert.Contains(t, out, "random:")
	assert.Contains(t, out, extractRandomPartHexV7(u))
	assert.Contains(t, out, "2025-05-05 03:03:03 +0000 UTC")
}

func uuidFromBytes(b ...byte) uuid.UUID {
	if len(b) != 16 {
		panic(fmt.Sprintf("uuidFromBytes: got %d bytes, want 16", len(b)))
	}

	var u uuid.UUID
	copy(u[:], b)

	return u
}
