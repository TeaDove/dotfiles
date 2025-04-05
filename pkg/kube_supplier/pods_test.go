package kube_supplier

import (
	"github.com/stretchr/testify/require"
	"github.com/teadove/teasutils/utils/test_utils"
	"testing"
)

func TestIntegration_KubeSupplier_GetContainers_Ok(t *testing.T) {
	supplier, err := NewSupplier()
	require.NoError(t, err)

	ctx := test_utils.GetLoggedContext()
	containers, err := supplier.GetContainerInfo(ctx)
	require.NoError(t, err)

	test_utils.LogAny(containers)
}
