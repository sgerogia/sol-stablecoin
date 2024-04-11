const { encryptEth, getEncryptionKey } = require("../util")

task("5-auth-granted", "Triggers an authGranted for PGBP")
  .addParam("contract", "The PGBP contract address")
  .addParam("requestId", "The current mint requestId")
  .addParam("consentCode", "The consent code as returned by the OpenBanking consent approval redirection")
  .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const networkId = network.name
    const consentCode = taskArgs.consentCode
    const requestId = taskArgs.requestId
    const privateKey = process.env.PRIVATE_KEY

    console.log("Initiating authGranted for PGBP (contract", contractAddr, ") on network", networkId)
    const ProvableGBP = await ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const [signer] = await ethers.getSigners()

    //Create connection to Contract and call the getter function
    const gbpContract = new ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    // Get contract owner's public encryption key
    const serverPubKey = ethers.utils.toUtf8String(await gbpContract.publicKey())
    // console.log("Server pub. key", serverPubKey)

    // Get our own public encryption key
    const myPubKey = getEncryptionKey(privateKey)

    const payload = {
        consentCode: consentCode,
        publicKey: myPubKey
    }
    const encryptedData = await encryptEth(serverPubKey, payload)

    console.log("Encrypted data", encryptedData)

    const trx = await gbpContract.authGranted(requestId, ethers.utils.toUtf8Bytes(encryptedData))
    console.log("Done. Trx:", trx.hash)
  })

module.exports = {}