// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";

contract ProvableGBP is ERC20, ERC20Burnable, Pausable, Ownable {

    using SafeMath for uint256;

    // --- Constants ---
    // Requests are considered expired after 2h
    uint256 public constant getExpiryTime = 2 hours;
    // % of the trx we keep in bips (10 = 0.1%)
    uint256 public constant seignorageFee = 10;
    // 100% in bips
    uint256 public constant oneHundredPercent = 10000;
    // the actual amount we are going to mint
    uint256 public constant actualMintedPercentage = oneHundredPercent - seignorageFee;

    // the public key of the server-side account
    bytes public publicKey = "";

    struct Commitment {
        bytes31 paramsHash;
        address requester;
        uint256 expiration;
        uint256 amount;
    }

    // map of mint commitments
    mapping(bytes32 => Commitment) private s_mintCommitments;


    constructor(bytes memory _publicKey) ERC20("Provable GBP", "PGBP") {
        publicKey = _publicKey;
    }

    event MintRequest(
        address indexed requester,
        bytes32 indexed requestId,
        uint256 amount,
        uint256 expiration,
        bytes encryptedData
    );

    event AuthRequest(
        address indexed requester,
        bytes32 indexed requestId,
        bytes authEncryptedData
    );

    event AuthGranted(
        address indexed requester,
        bytes32 indexed requestId,
        bytes grantEncryptedData
    );

    function pause() public onlyOwner {
        _pause();
    }

    function unpause() public onlyOwner {
        _unpause();
    }

    function setPublicKey(bytes memory _publicKey) public onlyOwner {
        publicKey = _publicKey;
    }

    /**
     * @notice Not implemented, cannot mint directly.
     */
    function mint(address, uint256) view public onlyOwner {
        // nameless function params mute "unused param" compiler warnings
        revert("You cannot mint directly");
    }

    function _beforeTokenTransfer(address from, address to, uint256 amount)
    internal
    whenNotPaused
    override
    {
        super._beforeTokenTransfer(from, to, amount);
    }

    /**
     * @notice Creates an internal mint Commitment and emits a MintRequest event. Pausable.
     * @param amount The amount to be minted (specified in 10^18 decimals)
     * @param encryptedData The encrypted payload of the request
     */
    function mintRequest(uint256 amount, bytes calldata encryptedData)
    public
    whenNotPaused {

        // TODO: add checks here, e.g. duplicate request, too many in the queue,...
        (bytes32 requestId, uint256 expiration) = _processMintRequest(
            msg.sender,
            amount,
            encryptedData
        );
        emit MintRequest(msg.sender, requestId, amount, expiration, encryptedData);
    }

    /**
     * @notice Triggered by the owner, emits an AuthRequest event for the original requester. Pausable.
     * @param requestId the original mint request id
     * @param serverEncryptedData the current auth. request's encrypted data
     */
    function authRequest(
        bytes32 requestId,
        bytes calldata serverEncryptedData
    )
    public
    onlyOwner
    whenNotPaused
    validateRequestId(requestId)
    validateNotExpired(requestId) {

        emit AuthRequest(s_mintCommitments[requestId].requester, requestId, serverEncryptedData);
    }


    /**
    * @notice Triggered by the original requester, emits an AuthGranted event for the server. Pausable.
     * @param requestId the original mint request id
     * @param encryptedData the grant's encrypted data
     */
    function authGranted(bytes32 requestId, bytes calldata encryptedData)
    public
    whenNotPaused
    validateRequestId(requestId)
    validateSameRequester(requestId, msg.sender) {

        // TODO: does it make sense to check for expiry? or not since this THE final step?

        emit AuthGranted(msg.sender, requestId, encryptedData);
    }

    /**
     * @notice Triggered by the owner, when the fiat funds have cleared. Does the mint, minus seignorage. Pausable.
     * @param requestId the original mint request id
     */
    function paymentComplete(bytes32 requestId)
    public
    onlyOwner
    whenNotPaused
    validateRequestId(requestId) {

        // TODO: maybe add a lifecycle status?

        // get values
        uint256 amount = s_mintCommitments[requestId].amount;
        address receiver = s_mintCommitments[requestId].requester;

        // delete commitment
        delete s_mintCommitments[requestId];

        // do the mint
        _mint(receiver, amount.mul(actualMintedPercentage).div(oneHundredPercent));
    }

    function _processMintRequest(address sender, uint256 amount, bytes calldata encryptedData)
    internal
    returns (bytes32 requestId, uint256 expiration) {
        requestId = keccak256(abi.encodePacked(sender, amount, encryptedData));
        require(s_mintCommitments[requestId].paramsHash == 0, "Request appears to be a duplicate");
        // solhint-disable-next-line not-rely-on-time
        expiration = block.timestamp.add(getExpiryTime);
        bytes31 paramsHash = _buildParamsHash(amount, encryptedData, expiration);
        s_mintCommitments[requestId] = Commitment(paramsHash, sender, expiration, amount);
        return (requestId, expiration);
    }

    /**
     * @notice Build the bytes31 hash from the amount, encryptedData and expiration.
     * @param amount The amount to be minted (specified in 10^18 decimals)
     * @param encryptedData The encrypted payload of the request
     * @param expiration The expiration before the commitment becomes eligible for cleanup
     * @return hash bytes31
     */
    function _buildParamsHash(
        uint256 amount,
        bytes calldata encryptedData,
        uint256 expiration
    ) internal pure returns (bytes31) {
        return bytes31(keccak256(abi.encodePacked(amount, encryptedData, expiration)));
    }

    /**
     * @dev Reverts if request ID does not exist
     * @param requestId The given request ID to check in stored `commitments`
     */
    modifier validateRequestId(bytes32 requestId) {
        require(s_mintCommitments[requestId].paramsHash != 0, "Must have a valid requestId");
        _;
    }

    /**
     * @dev Reverts if the commitment identified by the request ID has an expiry in the past
     * @param requestId The given request ID to check in stored `commitments`
     */
    modifier validateNotExpired(bytes32 requestId) {
        // solhint-disable-next-line not-rely-on-time
        require(s_mintCommitments[requestId].expiration >= block.timestamp, "Request is expired");
        _;
    }

    /**
     * @dev Reverts if the commitment has a recorded address different to the requester
     * @param requester The account making the request
     */
    modifier validateSameRequester(bytes32 requestId, address requester) {
        require(s_mintCommitments[requestId].requester == requester, "Requester does not match");
        _;
    }
}
