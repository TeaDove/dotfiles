package cli

import (
	"context"
	"crypto/rand"
	"dotfiles/pkg/cli/utils"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

func CommandUUID(_ context.Context, command *cli.Command) error {
	printStrings(command, func() string {
		return uuid.New().String()
	})

	return nil
}

func CommandUUID7(_ context.Context, cmd *cli.Command) error {
	printStrings(cmd, func() string {
		u := uuid.Must(uuid.NewV7())

		if !cmd.Bool(verboseFlag.Name) {
			return u.String()
		}

		return verboseUUID(u)
	})

	return nil
}

var (
	maxFlag = &cli.BoolFlag{Name: "max", Usage: "set uuid random to max bits"} //nolint:gochecknoglobals // is ok
	minFlag = &cli.BoolFlag{Name: "min", Usage: "set uuid random to min bits"} //nolint:gochecknoglobals // is ok)
)

func CommandUUID7Time(_ context.Context, cmd *cli.Command) error {
	const timeLayout = "2006-01-02T15:04:05"

	uuidTime, err := time.Parse(timeLayout, cmd.Args().First())
	if err != nil {
		return errors.Wrapf(err, "parse, required format: %s, passed: %s", timeLayout, cmd.Args().First())
	}

	var u uuid.UUID

	switch {
	case cmd.Bool(maxFlag.Name):
		u = setV7time(uuid.Max, uuidTime)
	case cmd.Bool(minFlag.Name):
		u = setV7time(uuid.Nil, uuidTime)
	default:
		u = setV7time(uuid.New(), uuidTime)
	}

	if cmd.Bool(verboseFlag.Name) {
		fmt.Print(verboseUUID(u))
	} else {
		fmt.Print(u.String())
	}

	return nil
}

func CommandUUID7Decode(_ context.Context, cmd *cli.Command) error {
	text, err := utils.ReadFromPipeOrSTDIN()
	if err != nil {
		return errors.Wrap(err, "read from stdin or pipe")
	}

	if strings.TrimSpace(text) == "" {
		return errors.New("empty input")
	}

	u, err := uuid.Parse(strings.TrimSpace(text))
	if err != nil {
		return errors.Wrap(err, "uuid parse")
	}

	if cmd.Bool(verboseFlag.Name) {
		fmt.Print(verboseUUID(u))
	} else {
		t, r := time.Unix(u.Time().UnixTime()), extractRandomPartHexV7(u)
		fmt.Printf("%s %s\n", t.String(), r)
	}

	return nil
}

func CommandText(_ context.Context, command *cli.Command) error {
	printStrings(command, rand.Text)

	return nil
}

func setV7time(u uuid.UUID, t time.Time) uuid.UUID {
	milli := t.UnixMilli()
	seq := (t.UnixNano() - milli*1_000_000) >> 8

	u[0] = byte(milli >> 40)
	u[1] = byte(milli >> 32)
	u[2] = byte(milli >> 24)
	u[3] = byte(milli >> 16)
	u[4] = byte(milli >> 8)
	u[5] = byte(milli)
	u[6] = 0x70 | byte(seq>>8)&0x0F
	u[7] = byte(seq)

	return u
}

func extractRandomPartHexV7(u uuid.UUID) string {
	randA := (uint16(u[6]&0x0F) << 8) | uint16(u[7])
	randB := uint64(u[8]&0x3F)<<56 |
		uint64(u[9])<<48 |
		uint64(u[10])<<40 |
		uint64(u[11])<<32 |
		uint64(u[12])<<24 |
		uint64(u[13])<<16 |
		uint64(u[14])<<8 |
		uint64(u[15])
	x := new(big.Int).SetUint64(randB)
	x.Or(x, new(big.Int).Lsh(big.NewInt(int64(randA)), 62))

	return fmt.Sprintf("%019x", x)
}

func verboseUUID(u uuid.UUID) string {
	t, r := time.Unix(u.Time().UnixTime()), extractRandomPartHexV7(u)

	var out strings.Builder
	fmt.Fprintf(&out, "input: %s\n", color.CyanString(u.String()))
	fmt.Fprintf(&out, "time: %s\n", color.BlueString(t.String()))
	fmt.Fprintf(&out, "time utc: %s\n", color.BlueString(t.UTC().String()))
	fmt.Fprintf(&out, "random: %s\n", color.GreenString(r))

	return out.String()
}

func printStrings(cmd *cli.Command, fn func() string) {
	count, err := strconv.Atoi(cmd.Args().First())
	if err != nil || count <= 1 {
		fmt.Print(fn())

		return
	}

	var v = make([]string, 0, count)
	for range count {
		v = append(v, fn())
	}

	sep := " "
	if cmd.Bool(verboseFlag.Name) {
		sep = "\n"
	}

	fmt.Print(strings.Join(v, sep)) //nolint:forbidigo // is ok
}
