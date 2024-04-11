const { encryptEth } = require("../util")

task("3-auth-request", "Triggers an authRequest for PGBP, in response to a mintRequest")
  .addParam("contract", "The PGBP contract address")
  .addParam("requestId", "The original mint-request requestId")
  .addParam("url", "The user authentication url to encrypt")
  .addParam("consentId", "The user consent identifier to encrypt")
  .addParam("publicKey", "The receiver's public encryption key, as read from get-mint-request")
  .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const networkId = network.name
    const requestId = taskArgs.requestId
    const url = taskArgs.url
    const consentId = taskArgs.consentId
    const publicKey = taskArgs.publicKey

    console.log("Initiating authRequest for RequestId", requestId, "PGBP (contract", contractAddr, ") on network", networkId)
    const ProvableGBP = await ethers.getContractFactory("ProvableGBP")

    // Get signer information
    const [ signer ] = await ethers.getSigners()

    //Create connection to Contract
    const gbpContract = new ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    // encrypt method payload
    const payload = {
        url: url,
        consentId, consentId,
    }
    const encryptedData = await encryptEth(publicKey, payload)

    // ...and call method
    const trx = await gbpContract.authRequest(requestId, ethers.utils.toUtf8Bytes(encryptedData))
    console.log("Done. Trx:", trx.hash)
  })

module.exports = {}