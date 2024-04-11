const { ethers } = require("ethers");
const { encrypt, decrypt } = require('@metamask/eth-sig-util');
const { stripHexPrefix } = require('ethereumjs-util');
const { getEncryptionPublicKey } = require("@metamask/eth-sig-util")

const ALGO = 'x25519-xsalsa20-poly1305';

/*
 * Performs the encryption using eth-sig-util.encrypt.
 * @param publicKey a key in hex (0x...) or base64
 * @param dataStructToEncrypt an object, which will be converted to JSON and encrypted
 */
const encryptEth = async (publicKey, dataStructToEncrypt) => {

    const pKey = publicKey.startsWith('0x')
            ? Buffer.from(stripHexPrefix(publicKey), 'hex').toString('base64')
            : publicKey;

    const dataToEncrypt = JSON.stringify(dataStructToEncrypt);

    const encr = JSON.stringify(
        encrypt({
            publicKey: pKey,
            data: dataToEncrypt,
            version: ALGO,
        })
    );
    return encr;
}

/*
 * Performs the decryption using the private key.
 * @param privateKey in hex (0x...) or base64 format
 * @param encrData a string representation of the eth-sig-util data structure
 * {
 *   version: 'x25519-xsalsa20-poly1305',
 *   nonce: <base64>,
 *   ephemPublicKey: <base64>,
 *   ciphertext: <base64>
 * }
 */
const decryptEth = async (privateKey, encrData) => {

    const data = stripHexPrefix(encrData);
    const encrJson = JSON.parse(encrData);

    const pKey = privateKey.startsWith('0x')
            ? stripHexPrefix(privateKey)
            : Buffer.from(privateKey, 'base64').toString('hex');

    const decr = await decrypt({
        encryptedData: encrJson,
        privateKey: pKey,
    });
    return decr;
}

/*
 * Utility wrapper around eth-sig-util.getEncryptionPublicKey. Takes care of key format conversions.
 * @param privateKey in hex or base64 format
 */
const getEncryptionKey = (privateKey) => {

    const pKey = privateKey.startsWith('0x')
            ? stripHexPrefix(privateKey)
            : Buffer.from(privateKey, 'base64').toString('hex');

    return getEncryptionPublicKey(pKey);
}

module.exports = {
    encryptEth,
    decryptEth,
    getEncryptionKey
}
