const { encryptEth, getEncryptionKey } = require("../util")

task("1-mint-request", "Triggers a mintRequest for PGBP")
  .addParam("contract", "The PGBP contract address")
  .addParam("amount", "The amount of PGBP to mint (natural number, not wei)")
  .addParam("institutionId", "The institution id as defined in the Yapily API")
  .addParam("sortCode", "The account holder's bank account sort code")
  .addParam("accountNumber", "The account holder's bank account number")
  .addParam("name", "The account holder's full name")
  .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const networkId = network.name
    const amount = taskArgs.amount
    const institutionId = taskArgs.institutionId
    const sortCode = taskArgs.sortCode
    const accountNumber = taskArgs.accountNumber
    const name = taskArgs.name
    const privateKey = process.env.PRIVATE_KEY

    console.log("Initiating mintRequest for", amount, "PGBP (contract", contractAddr, ") on network", networkId)
    const ProvableGBP = await ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const [signer] = await ethers.getSigners()

    //Create connection to Contract and call the getter function
    const gbpContract = new ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    // Get contract owner's public encryption key
    const serverPubKey = ethers.utils.toUtf8String(await gbpContract.publicKey())
//    console.log("Server pub. key", serverPubKey)

    // Get our own public encryption key
    const myPubKey = getEncryptionKey(privateKey)

    const payload = {
        institutionId: institutionId,
        sortCode: sortCode,
        accountNumber: accountNumber,
        name: name,
        publicKey: myPubKey
    }
    const encryptedData = await encryptEth(serverPubKey, payload)

    console.log("Encrypted data", encryptedData)

    const trx = await gbpContract.mintRequest(ethers.utils.parseEther(amount), ethers.utils.toUtf8Bytes(encryptedData))
    console.log("Done. Trx:", trx.hash)
  })

module.exports = {}