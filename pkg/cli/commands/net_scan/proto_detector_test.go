package net_scan

import (
	"testing"

	"github.com/teadove/teasutils/utils/test_utils"
)

func TestProtoDetector(t *testing.T) {
	t.Parallel()

	r := New()

	test_utils.Pprint(r.protoDetection(test_utils.GetLoggedContext(), "192.168.0.1", 80))
	test_utils.Pprint(r.protoDetection(test_utils.GetLoggedContext(), "70.34.196.45", 22))
	test_utils.Pprint(r.protoDetection(test_utils.GetLoggedContext(), "192.168.0.166", 5000))
	test_utils.Pprint(r.protoDetection(test_utils.GetLoggedContext(), "192.168.0.113", 1961))
	test_utils.Pprint(r.protoDetection(test_utils.GetLoggedContext(), "70.34.196.45", 8080))
}
