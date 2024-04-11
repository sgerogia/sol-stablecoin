const { ethers } = require("ethers")

task("balance", "Prints an account's balance in PGBP")
  .addParam("contract", "The PGBP contract address")
  .addParam("account", "The balance holder's address")
  .setAction(async (taskArgs) => {

    const account = hre.ethers.utils.getAddress(taskArgs.account)
    const contractAddr = taskArgs.contract
    const networkId = network.name

    console.log("Fetching PGBP", contractAddr, "balance for", account, "on network", networkId)
    const ProvableGBP = await hre.ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const accounts = await hre.ethers.getSigners()
    const signer = accounts[0]

    //Create connection to Contract and call the getter function
    const gbpContract = new hre.ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    let value = await gbpContract.balanceOf(account)
    console.log(ethers.utils.formatEther(balance), "PGBP")
  })

module.exports = {}
