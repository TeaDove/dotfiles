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

func CommandUUID7(_ context.Context, command *cli.Command) error {
	printStrings(command, func() string {
		return uuid.Must(uuid.NewV7()).String()
	})

	return nil
}

func randomPartHexV7(u uuid.UUID) string {
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

func CommandUUID7Decode(_ context.Context, cmd *cli.Command) error {
	text, err := utils.ReadFromPipeOrSTDIN()
	if err != nil {
		return errors.Wrap(err, "read from stdin or pipe")
	}

	u, err := uuid.Parse(strings.TrimSpace(text))
	if err != nil {
		return errors.Wrap(err, "uuid parse")
	}

	t, r := time.Unix(u.Time().UnixTime()), randomPartHexV7(u)

	if cmd.Bool(verboseFlag.Name) {
		fmt.Printf("input: %s\n", color.CyanString(u.String()))
		fmt.Printf("time: %s\n", color.BlueString(t.String()))
		fmt.Printf("time utc: %s\n", color.BlueString(t.UTC().String()))
		fmt.Printf("random: %s\n", color.GreenString(r))
	} else {
		fmt.Printf("%s %s\n", t.String(), r)
	}

	return nil
}

func CommandText(_ context.Context, command *cli.Command) error {
	printStrings(command, rand.Text)

	return nil
}

func printStrings(command *cli.Command, fn func() string) {
	count, err := strconv.Atoi(command.Args().First())
	if err != nil || count <= 1 {
		fmt.Print(fn())

		return
	}

	var v = make([]string, 0, count)
	for range count {
		v = append(v, fn())
	}

	fmt.Print(strings.Join(v, " ")) //nolint:forbidigo // is ok
}
