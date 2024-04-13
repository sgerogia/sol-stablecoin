import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "hardhat-deploy";
import "@nomiclabs/hardhat-ethers";
import "@solidstate/hardhat-bytecode-exporter";
import "./tasks"

// Infura API
const API_KEY = process.env.INFURA_TOKEN
// ETH
const MAINNET_RPC_URL = "https://mainnet.infura.io/v3/" + API_KEY
const GOERLI_RPC_URL = "https://goerli.infura.io/v3/" + API_KEY
const SEPOLIA_RPC_URL = "https://sepolia.infura.io/v3/" + API_KEY
// NeonEVM
const NEON_DEVNET_RPC_URL = "https://devnet.neonevm.org"
const NEON_MAINNET_RPC_URL = "https://neon-proxy-mainnet.solana.p2p.org"


const PRIVATE_KEY = process.env.PRIVATE_KEY
// optional
const MNEMONIC = process.env.MNEMONIC || "Your mnemonic"
const FORKING_BLOCK_NUMBER = process.env.FORKING_BLOCK_NUMBER !== undefined ? parseInt(process.env.FORKING_BLOCK_NUMBER) : 0

// Your API key for Etherscan, obtain one at https://etherscan.io/
const ETHERSCAN_API_KEY = process.env.ETHERSCAN_API_KEY || "Your etherscan API key"
const REPORT_GAS = process.env.REPORT_GAS || false

const config: HardhatUserConfig = {
    defaultNetwork: "hardhat",
    networks: {
        hardhat: {
            // If you want to do some forking set `enabled` to true
            forking: {
                url: MAINNET_RPC_URL,
                blockNumber: FORKING_BLOCK_NUMBER,
                enabled: false,
            },
            chainId: 31337,
        },
        ganache: {
            accounts: PRIVATE_KEY !== undefined ? [PRIVATE_KEY] : [],
            url: "http://127.0.0.1:8545",
            chainId: 1337,
            gas: 2100000,
            gasPrice: 600000000,
        },

        // NeonEVM live networks
        neondevnet: { 
            url: NEON_DEVNET_RPC_URL,
            accounts: PRIVATE_KEY !== undefined ? [PRIVATE_KEY] : [],
            saveDeployments: true,
            chainId: 245022926,
        },
        neonmainnet: {
            url: NEON_MAINNET_RPC_URL,
            accounts: PRIVATE_KEY !== undefined ? [PRIVATE_KEY] : [],
            saveDeployments: true,
            chainId: 245022934,
        },

        // Ethereum live networks
        goerli: { // testnet
            url: GOERLI_RPC_URL,
            accounts: PRIVATE_KEY !== undefined ? [PRIVATE_KEY] : [],
            //   accounts: {
            //     mnemonic: MNEMONIC,
            //   },
            saveDeployments: true,
            chainId: 5,
            //gas: 2000000,
            //gasPrice: 20000000000, // affects how quickly (if at all) the trx goes in. Needs fine-tuning for mainnet
            //gasMultiplier: 1.4
        },
        sepolia: { // testnet
            url: SEPOLIA_RPC_URL,
            accounts: PRIVATE_KEY !== undefined ? [PRIVATE_KEY] : [],
            //   accounts: {
            //     mnemonic: MNEMONIC,
            //   },
            saveDeployments: true,
            chainId: 11155111,
            //gas: 2000000,
            //gasPrice: 20000000000, // affects how quickly (if at all) the trx goes in. Needs fine-tuning for mainnet
            //gasMultiplier: 1.4
        },
        mainnet: {
            url: MAINNET_RPC_URL,
            accounts: PRIVATE_KEY !== undefined ? [PRIVATE_KEY] : [],
            //   accounts: {
            //     mnemonic: MNEMONIC,
            //   },
            saveDeployments: true,
            chainId: 1,
        },
    },
    etherscan: {
        // npx hardhat verify --network <NETWORK> <CONTRACT_ADDRESS> <CONSTRUCTOR_PARAMETERS>
        apiKey: {
            goerli: ETHERSCAN_API_KEY,
            neonevm: "test"
        },
        customChains: [
            {
                network: "neonevm",
                chainId: 245022926,
                urls: {
                    apiURL: "https://devnet-api.neonscan.org/hardhat/verify",
                    browserURL: "https://devnet.neonscan.org"
                }
            },
            {
                network: "neonevm",
                chainId: 245022934,
                urls: {
                    apiURL: "https://api.neonscan.org/hardhat/verify",
                    browserURL: "https://neonscan.org"
                }
            }
        ] 
    },
    gasReporter: {
        enabled: REPORT_GAS,
        currency: "USD",
        outputFile: "gas-report.txt",
        noColors: true,
        // coinmarketcap: process.env.COINMARKETCAP_API_KEY,
    },
    contractSizer: {
        runOnCompile: false,
        only: ["ProvableGBP"],
    },
    namedAccounts: {
        deployer: {
            default: 0, // here this will by default take the first account as deployer
            1: 0, // similarly on mainnet it will take the first account as deployer. Note though that depending on how hardhat network are configured, the account 0 on one network can be different than on another
        },
        user: {
            default: 1,
        },
    },
    solidity: "0.8.17",

    bytecodeExporter: {
      path: './artifacts/contracts',
      runOnCompile: true,
      clear: true,
      flat: true,
    }
};

export default config;
