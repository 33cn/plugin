package offline

import "github.com/spf13/cobra"

const gasLimit uint64 = 10000 * 800

func EthOfflineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eth",
		Short: "create and sign offline tx to deploy and set dex contracts to eth",
		Args:  cobra.MinimumNArgs(1),
	}
	var query = new(queryCmd)
	var signdeployCmd = new(SignCmd)
	var deploy = new(DeployContract)
	var addpool = new(AddPool)
	var update = new(updateAllocPoint)
	var transOwner = new(transferOwnerShip)
	cmd.AddCommand(
		query.queryCmd(),        //query fromAccount info such as: nonce,gasprice
		signdeployCmd.signCmd(), //sign fatory.weth9,pancakrouter contract
		addpool.AddPoolCmd(),    //call contract
		update.UpdateAllocPointCmd(),
		transOwner.TransferOwnerShipCmd(),
		deploy.DeployCmd(), //send singned tx to deploy contract:factory,weth9,pancakerouter.
	)
	return cmd
}
