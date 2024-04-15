import { Chain } from '@usedapp/core'
import { ethers } from 'ethers'

/* Hack: Copied from @usedapp/core, as they are not exported  */
export const getAddressLink = (explorerUrl: string) => (address: string) => `${explorerUrl}/address/${address}`
export const getTransactionLink = (explorerUrl: string) => (txnId: string) => `${explorerUrl}/tx/${txnId}`

const neonExplorerUrl = 'https://neonscan.org/'


export const Neon: Chain = {
  chainId: 245022934,
  chainName: 'NeonEVM',
  isTestChain: false,
  isLocalChain: false,
  multicallAddress: '0xca11bde05977b3631167028862be2a173976ca11',
  rpcUrl: 'https://neon-proxy-mainnet.solana.p2p.org/rpc',
  nativeCurrency: {
    name: 'NEON',
    symbol: 'NEON',
    decimals: 18,
  },
  blockExplorerUrl: neonExplorerUrl,
  getExplorerAddressLink: getAddressLink(neonExplorerUrl),
  getExplorerTransactionLink: getTransactionLink(neonExplorerUrl),
}

const neonDevnetExplorerUrl = 'https://devnet.neonscan.org/'

export const NeonDevnet: Chain = {
  chainId: 245022926,
  chainName: 'Neon Devnet',
  isTestChain: true,
  isLocalChain: false,
  multicallAddress: '0xcA11bde05977b3631167028862bE2a173976CA11',
  rpcUrl: 'https://devnet.neonevm.org/rpc',
  nativeCurrency: {
    name: 'NEON',
    symbol: 'NEON',
    decimals: 18,
  },
  blockExplorerUrl: neonDevnetExplorerUrl,
  getExplorerAddressLink: getAddressLink(neonDevnetExplorerUrl),
  getExplorerTransactionLink: getTransactionLink(neonDevnetExplorerUrl),
}

export const NeonProvider = new ethers.providers.JsonRpcProvider(Neon.rpcUrl)
export const NeonDevnetProvider = new ethers.providers.JsonRpcProvider(NeonDevnet.rpcUrl)