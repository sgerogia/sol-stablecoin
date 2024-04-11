const { decryptEth } = require("../util")

task("4-get-auth-request", "Receive the latest AuthRequest message for PGBP")
    .addParam("contract", "The PGBP contract address")
    .addParam("account", "The payer's account address which initiated the flow (used for filtering)")
    .addOptionalParam("poll", "Should we poll until we get a new event? [true/false] Only use on a public net. Default to false", "false")
    .setAction(async (taskArgs) => {

    const contractAddr = taskArgs.contract
    const account = taskArgs.account
    const networkId = network.name
    const poll = "true" === taskArgs.poll

    console.log("Listening for AuthRequest events from PGBP (contract", contractAddr, ") on network", networkId)
    const ProvableGBP = await ethers.getContractFactory("ProvableGBP")

    //Get signer information
    const signer = new ethers.Wallet(process.env.PRIVATE_KEY, ethers.provider)

    //Create connection to Contract and filter for events
    const gbpContract = new ethers.Contract(contractAddr, ProvableGBP.interface, signer)

    const eventFilter = gbpContract.filters.AuthRequest(account)
    let events = await gbpContract.queryFilter(eventFilter)
    // different behaviour if polling or not
    if (poll) {
        let l = events.length
        console.log("Polling the contract for new events. Press Ctr+C to stop...")
        while (true) {
            await sleep(2000)
            events = await gbpContract.queryFilter(eventFilter)
            if (l != events.length) {
                break;
            }
            console.log("Im Westen nichts Neues...")
        }
    } else {
        // exit early
        if (events.length == 0) {
            console.log("No AuthRequest events yet. Try again in a few minutes...")
            return
        } else {
            console.log("Found", events.length, "events. Showing the latest.")
        }
    }

    const event = events[events.length-1].args
    console.log("AuthRequest received");
    console.log("\tRequestId", event.requestId);
    console.log("\tRequester", event.requester);
    console.log("\tEncr. data", ethers.utils.toUtf8String(event.authEncryptedData));

    const decryptedData = await decryptEth(signer.privateKey, ethers.utils.toUtf8String(event.authEncryptedData))
    const rawData = JSON.parse(decryptedData)

    const url = rawData.url
    const consentId = rawData.consentId

    console.log("--------------------------------------------------")
    console.log("Use the following URL to grant the consent")
    console.log("URL:", url)
    console.log("Consent:", consentId)
  })

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
module.exports = {}