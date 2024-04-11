import { time, loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { expect, assert } from "chai";
import { ethers } from "hardhat";
import { encryptEth, decryptEth, getEncryptionKey } from "../tasks/util";

describe("ProvableGBP - mintRequest encryption", function () {

    const payload = {
            sortCode: "400101",
            accountNumber: "20304050",
            name: "King Charles the Unemployed",
            publicKey: "",
    }

    async function deployFixture() {

        // default HH account
        const [ defaultHardhatAccount ] = await ethers.getSigners();

        // our own accounts
        const owner = ethers.Wallet.createRandom().connect(ethers.provider);
        const otherAccount = ethers.Wallet.createRandom().connect(ethers.provider);

        // update the public key of the other account in the payload
        payload.publicKey = getEncryptionKey(otherAccount.privateKey);
//        console.log("Encryption key", payload.publicKey);

        // fund our own accounts
        await defaultHardhatAccount.sendTransaction({to: owner.address, value: ethers.utils.parseEther('10')});
        await defaultHardhatAccount.sendTransaction({to: otherAccount.address, value: ethers.utils.parseEther('10')});

        // deploy from our own accounts
        const ProvableGBP = await ethers.getContractFactory("ProvableGBP", owner);
        const gbp = await ProvableGBP.deploy(ethers.utils.toUtf8Bytes(getEncryptionKey(owner.privateKey)));

        return { gbp, owner, otherAccount };
    }

    describe("Happy path", function () {

        it("Should encrypt & decrypt the payload using the contract's public key", async function () {

            // arrange
            const { gbp, owner, otherAccount } = await loadFixture(deployFixture);

            const ownerPublicKey = ethers.utils.toUtf8String(await await gbp.connect(otherAccount).publicKey());
//            console.log("Owner public key", ownerPublicKey);

            const mintAmount = 100000000;
            const encryptedData = await encryptEth(ownerPublicKey, payload);

//            console.log("Encrypted", encryptedData);

            // act & assert
            await gbp.connect(otherAccount).mintRequest(mintAmount, ethers.utils.toUtf8Bytes(encryptedData));

            // assert
            const eventTrx = (await gbp.queryFilter("MintRequest"))[0];
            const event = eventTrx.args;
            console.log("MintRequest received");

            const decryptedData = await decryptEth(owner.privateKey, ethers.utils.toUtf8String(event.encryptedData));

            console.log("Decrypted", decryptedData);
            expect(JSON.parse(decryptedData)).to.deep.equal(payload);
        });
    });
});
