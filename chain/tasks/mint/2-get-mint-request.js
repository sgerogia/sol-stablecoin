const { decryptEth } = require("../util")

task("2-get-mint-request", "Receive the latest MintRequest message for PGBP")
    .addParam("contract", "The PGBP contract address")
    .addParam("account", "The payer's account address which initiated the flow (used for filtering)")
    .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const account = ethers.utils.getAddress(taskArgs.account)
    const networkId = network.name

    console.log("Listening for MintRequest for PGBP (contract", contractAddr, ") on network", networkId)
    const ProvableGBP = await ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const signer = new ethers.Wallet(process.env.PRIVATE_KEY, ethers.provider)

    //Create connection to Contract and filter for events
    const gbpContract = new ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    const eventFilter = gbpContract.filters.MintRequest(account)
    const events = await gbpContract.queryFilter(eventFilter)
    // exit early
    if (events.length == 0) {
        console.log("No MintRequest events yet. Try again in a few minutes...")
        return
    } else {
        console.log("Found", events.length, "MintRequest events. Showing the last one.")
    }

    const event = events[events.length-1].args
    console.log("MintRequest received");
    console.log("\tRequestId", event.requestId);
    console.log("\tRequester", event.requester);
    console.log("\tAmount", event.amount);
    console.log("\tExpiration", event.expiration);
    console.log("\tEncr. data", ethers.utils.toUtf8String(event.encryptedData));

    const decryptedData = await decryptEth(signer.privateKey, ethers.utils.toUtf8String(event.encryptedData))
    const rawData = JSON.parse(decryptedData)

    const requester = event.requester
    const amount = ethers.utils.formatEther(event.amount)
    const institutionId = rawData.institutionId
    const name = rawData.name
    const sortCode = rawData.sortCode
    const accountNumber = rawData.accountNumber
    const publicKey = rawData.publicKey

    console.log("--------------------------------------------------")
    console.log("MintRequest encr. payload");
    console.log("\tInstitutionId", institutionId);
    console.log("\tSort code", sortCode);
    console.log("\tAcc. number", accountNumber);
    console.log("\tName", name);
    console.log("\tPublic key", publicKey);

  })

module.exports = {}