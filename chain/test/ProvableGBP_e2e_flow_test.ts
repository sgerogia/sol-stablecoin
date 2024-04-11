import { time, loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { expect, assert } from "chai";
import { ethers } from "hardhat";

describe("ProvableGBP - E2E flow", function () {

  const publicKey = "publicKey";

  async function deployFixture() {

    // Contracts are deployed using the first signer/account by default
    const [owner, otherAccount] = await ethers.getSigners();

    const ProvableGBP = await ethers.getContractFactory("ProvableGBP");
    const gbp = await ProvableGBP.deploy(ethers.utils.toUtf8Bytes(publicKey));

    return { gbp, owner, otherAccount };
  }

  describe("Happy path", function () {

    it("Should transition mintRequest() -> MintRequest event", async function () {

      // arrange
      const { gbp, owner, otherAccount } = await loadFixture(deployFixture);

      const mintAmount = 100000000;
      const encrData = "Some data";

      // act & assert
      await gbp.connect(otherAccount).mintRequest(mintAmount, ethers.utils.toUtf8Bytes(encrData));

      // assert
      let event = (await gbp.queryFilter("MintRequest"))[0].args
      console.log("MintRequest received");
      expect(event.requestId).to.not.be.empty;
      assert.equal(event.requester, otherAccount.address);
      assert.equal(event.amount, mintAmount);
      expect(event.expiration).to.be.greaterThan(0);
      assert.equal(encrData, ethers.utils.toUtf8String(event.encryptedData));

    });

    it("Should transition authRequest() -> AuthRequest event", async function () {

      // arrange
      const { gbp, owner, otherAccount } = await loadFixture(deployFixture);

      const mintAmount = 100000000;
      const encrData = "Some data";
      const serverEncrData = "More data";

      await gbp.connect(otherAccount).mintRequest(mintAmount, ethers.utils.toUtf8Bytes(encrData));

      let ev1 = (await gbp.queryFilter("MintRequest"))[0].args
      const reqId = ev1.requestId;
      const exp = ev1.expiration;

      // act
      await gbp.authRequest(
        reqId,
        ethers.utils.toUtf8Bytes(serverEncrData));

      // assert
      let ev2 = (await gbp.queryFilter("AuthRequest"))[0].args
      console.log("AuthRequest received");
      assert.equal(ev2.requestId, reqId);
      assert.equal(ev2.requester, otherAccount.address);
      assert.equal(serverEncrData, ethers.utils.toUtf8String(ev2.authEncryptedData));
    });

    it("Should transition authGranted() -> AuthGranted event", async function () {

      // arrange
      const { gbp, owner, otherAccount } = await loadFixture(deployFixture);

      const mintAmount = 100000000;
      const encrData = "Some data";
      const serverEncrData = "More data";
      const authEncrData = "Auth data";

      await gbp.connect(otherAccount).mintRequest(mintAmount, ethers.utils.toUtf8Bytes(encrData));

      let ev1 = (await gbp.queryFilter("MintRequest"))[0].args
      const reqId = ev1.requestId;
      const exp = ev1.expiration;

      await gbp.authRequest(
        reqId,
        ethers.utils.toUtf8Bytes(serverEncrData));

      // act
      await gbp.connect(otherAccount).authGranted(reqId, ethers.utils.toUtf8Bytes(authEncrData));

      // assert
      let ev = (await gbp.queryFilter("AuthGranted"))[0].args
      console.log("AuthGranted received");
      assert.equal(ev.requestId, reqId);
      assert.equal(ev.requester, otherAccount.address);
      assert.equal(authEncrData, ethers.utils.toUtf8String(ev.grantEncryptedData));
    });

    it("Should transition paymentComplete() -> _mint()", async function () {

      // arrange
      const { gbp, owner, otherAccount } = await loadFixture(deployFixture);

      const mintAmount = 100000000;
      const encrData = "Some data";
      const serverEncrData = "More data";
      const authEncrData = "Auth data";

      await gbp.connect(otherAccount).mintRequest(mintAmount, ethers.utils.toUtf8Bytes(encrData));

      let ev1 = (await gbp.queryFilter("MintRequest"))[0].args
      const reqId = ev1.requestId;
      const exp = ev1.expiration;

      await gbp.authRequest(
        reqId,
        ethers.utils.toUtf8Bytes(serverEncrData));

      await gbp.connect(otherAccount).authGranted(reqId, ethers.utils.toUtf8Bytes(authEncrData));

      const mintPerc = await gbp.actualMintedPercentage();
      const hundPerc = await gbp.oneHundredPercent();

      // act & assert
      await expect(
              gbp.paymentComplete(reqId)
            ).to.changeTokenBalances(gbp, [otherAccount], [mintAmount * mintPerc / hundPerc]);
    });

  });


});
