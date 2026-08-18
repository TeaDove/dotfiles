package closeutils

import (
	"io"

	"github.com/cockroachdb/errors"
)

func MustClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		panic(errors.Wrap(err, "close"))
	}
}
