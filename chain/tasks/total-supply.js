const { ethers } = require("ethers")

task("total-supply", "Prints the total supply of PGBP")
  .addParam("contract", "The PGBP contract address")
  .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const networkId = network.name

    console.log("Fetching PGBP", contractAddr, "total supply on network", networkId)
    const ProvableGBP = await hre.ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const accounts = await hre.ethers.getSigners()
    const signer = accounts[0]

    //Create connection to Contract and call the getter function
    const gbpContract = new hre.ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    let value = await gbpContract.totalSupply()
    console.log(ethers.utils.formatEther(balance), "PGBP")
  })

module.exports = {}
