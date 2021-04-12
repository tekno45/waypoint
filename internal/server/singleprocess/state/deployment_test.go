package state

import (
	"testing"

	pb "github.com/hashicorp/waypoint/internal/server/gen"
	serverptypes "github.com/hashicorp/waypoint/internal/server/ptypes"
	"github.com/stretchr/testify/require"
)

func TestDeployment(t *testing.T) {
	deploymentOp.Test(t)
}

func TestDeploymentPrune(t *testing.T) {
	t.Run("prunes old records", func(t *testing.T) {
		require := require.New(t)

		s := TestState(t)
		defer s.Close()

		require.NoError(s.DeploymentPut(false, serverptypes.TestValidDeployment(t, &pb.Deployment{
			Id: "A",
		})))

		require.NoError(s.DeploymentPut(false, serverptypes.TestValidDeployment(t, &pb.Deployment{
			Id: "B",
		})))

		require.NoError(s.DeploymentPut(false, serverptypes.TestValidDeployment(t, &pb.Deployment{
			Id: "C",
		})))

		memTxn := s.inmem.Txn(true)
		defer memTxn.Abort()

		cnt, err := deploymentOp.pruneOld(memTxn, 2)
		require.NoError(err)

		memTxn.Commit()

		require.Equal(1, cnt)
		require.Equal(2, deploymentOp.indexedRecords)

		dep, err := s.DeploymentGet(&pb.Ref_Operation{
			Target: &pb.Ref_Operation_Id{
				Id: "A",
			},
		})

		require.Error(err)
		require.Nil(dep)
	})
}
