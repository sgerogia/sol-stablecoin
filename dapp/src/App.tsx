import React from 'react'
import ReactDOM from 'react-dom'

import { getDefaultProvider } from 'ethers'

import { DAppProvider, useEtherBalance, useEthers, Config, DEFAULT_SUPPORTED_CHAINS, Mainnet, Sepolia } from '@usedapp/core'
import { Neon, NeonDevnet, NeonDevnetProvider } from './util/neon-chain'
import { formatEther } from '@ethersproject/units'

const config: Config = {
  autoConnect: true,
  readOnlyChainId: NeonDevnet.chainId,
  readOnlyUrls: {
    [Mainnet.chainId]: 'https://mainnet.infura.io/v3/8381b69184e54ff38ee0a1ebb60e8e63',
    [Sepolia.chainId]: 'https://sepolia.infura.io/v3/8381b69184e54ff38ee0a1ebb60e8e63',
    // [Neon.chainId]: 'https://neon-proxy-mainnet.solana.p2p.org',
    // [NeonDevnet.chainId]: 'https://devnet.neonevm.org',
  },
  // networks: [...DEFAULT_SUPPORTED_CHAINS, Neon, NeonDevnet]
}

const ConnectButton = () => {
  const { account, deactivate, activateBrowserWallet } = useEthers()
  // 'account' being undefined means that we are not connected.
  if (account) return <button onClick={() => deactivate()}>Disconnect</button>
  else return <button onClick={async () => { 
    try {
      await activateBrowserWallet()
      console.log('Click')
    } catch(error) {
      console.error("Failed to connect:", error)
    }
  }}>Connect</button>
}

ReactDOM.render(
  <DAppProvider config={config}>
    <App />
  </DAppProvider>,
  document.getElementById('root')
)

function App() {
  const { account, chainId } = useEthers()
  const etherBalance = useEtherBalance(account)
  if (chainId && config.readOnlyUrls && !config.readOnlyUrls[chainId]) {
    return <p>Please use either Neon or Neon Devnet.</p>
  }

  return (
    <div>
      <ConnectButton />
      {etherBalance && (
        <div className="balance">
          <br />
          Address:
          <p className="bold">{account}</p>
          <br />
          Balance:
          <p className="bold">{formatEther(etherBalance)}</p>
        </div>
      )}
    </div>
  )
}

export default App