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
	containers, err := supplier.GetContainersInfo(ctx)
	require.NoError(t, err)

	test_utils.Pprint(containers)
}

func TestIntegration_KubeSupplier_GetDeployments_Ok(t *testing.T) {
	supplier, err := NewSupplier()
	require.NoError(t, err)

	ctx := test_utils.GetLoggedContext()
	containers, err := supplier.GetDeploymentInfo(ctx)
	require.NoError(t, err)

	test_utils.Pprint(containers)
}
