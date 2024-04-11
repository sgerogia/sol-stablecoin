const { ethers } = require("ethers")

task("public-key", "Prints an account's balance in PGBP")
  .addParam("contract", "The PGBP contract address")
  .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const networkId = network.name

    console.log("Fetching TPP public key from PGBP", contractAddr, "on network", networkId)
    const ProvableGBP = await hre.ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const accounts = await hre.ethers.getSigners()
    const signer = accounts[0]

    //Create connection to Contract and call the getter function
    const gbpContract = new hre.ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    let value = await gbpContract.publicKey()
    console.log("Public Key", ethers.utils.toUtf8String(value))
  })

module.exports = {}
