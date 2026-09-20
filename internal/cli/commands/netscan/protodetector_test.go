package netscan

import (
	"testing"

	"github.com/teadove/teasutils/utils/testutils"
)

func TestProtoDetector(t *testing.T) {
	t.Parallel()

	r := New()

	testutils.Debug(r.protoDetection(testutils.Context(), "192.168.0.1", 80))
	testutils.Debug(r.protoDetection(testutils.Context(), "70.34.196.45", 22))
	testutils.Debug(r.protoDetection(testutils.Context(), "192.168.0.166", 5000))
	testutils.Debug(r.protoDetection(testutils.Context(), "192.168.0.113", 1961))
	testutils.Debug(r.protoDetection(testutils.Context(), "70.34.196.45", 8080))
}
