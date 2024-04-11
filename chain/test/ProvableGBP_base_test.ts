import { time, loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";
import { expect } from "chai";
import { ethers } from "hardhat";

describe("ProvableGBP - Base", function () {

  const key = "key";

  async function deployFixture() {

    // Contracts are deployed using the first signer/account by default
    const [owner, otherAccount] = await ethers.getSigners();

    const ProvableGBP = await ethers.getContractFactory("ProvableGBP");
    const gbp = await ProvableGBP.deploy(ethers.utils.toUtf8Bytes(key));

    return { gbp, owner, otherAccount };
  }

  async function deployFixtureWithPause() {

    // Contracts are deployed using the first signer/account by default
    const { gbp, owner, otherAccount } = await deployFixture();

    await gbp.pause();

    expect(await gbp.paused()).to.equal(true);

    return { gbp, owner, otherAccount };
  }

  describe("Deployment", function () {
    it("Should have 'mint' disabled", async function () {
      const { gbp, owner } = await loadFixture(deployFixture);

      await expect(gbp.mint(owner.address, 1000)).to.be.revertedWith("You cannot mint directly");
    });

    it("Should have the right number of decimals", async function () {
      const { gbp } = await loadFixture(deployFixture);

      expect(await gbp.decimals()).to.equal(18);
    });

    it("Should have the right symbol", async function () {
      const { gbp } = await loadFixture(deployFixture);

      expect(await gbp.symbol()).to.equal("PGBP");
    });

    it("Should have the right name", async function () {
      const { gbp } = await loadFixture(deployFixture);

      expect(await gbp.name()).to.equal("Provable GBP");
    });

    it("Should have the right total supply", async function () {
      const { gbp } = await loadFixture(deployFixture);

      expect(await gbp.totalSupply()).to.equal(0);
    });

    it("Should have the right public key", async function () {
      const { gbp } = await loadFixture(deployFixture);

      expect(ethers.utils.toUtf8String(await gbp.publicKey())).to.equal(key);
    });
  });

  describe("Pausing", function () {
    it("mintRequest is paused", async function () {
      const { gbp } = await loadFixture(deployFixtureWithPause);

      await expect(gbp.mintRequest(1000, ethers.utils.toUtf8Bytes("Test"))).to.be.revertedWith("Pausable: paused");
    });

    it("authRequest is paused", async function () {
      const { gbp, owner } = await loadFixture(deployFixtureWithPause);

      await expect(gbp.authRequest(
        ethers.utils.toUtf8Bytes("29fa9aa13bf1468788b7cc4a500a45b8"),
        ethers.utils.toUtf8Bytes("serverEncryptedData")
      )).to.be.revertedWith("Pausable: paused");
    });

    it("authGranted is paused", async function () {
      const { gbp } = await loadFixture(deployFixtureWithPause);

      await expect(gbp.authGranted(
        ethers.utils.toUtf8Bytes("29fa9aa13bf1468788b7cc4a500a45b8"),
        ethers.utils.toUtf8Bytes("encryptedData")
      )).to.be.revertedWith("Pausable: paused");
    });

    it("paymentComplete is paused", async function () {
      const { gbp } = await loadFixture(deployFixtureWithPause);

      await expect(gbp.paymentComplete(
        ethers.utils.toUtf8Bytes("29fa9aa13bf1468788b7cc4a500a45b8"),
      )).to.be.revertedWith("Pausable: paused");
    });

  });

});
