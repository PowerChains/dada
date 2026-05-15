import { useState } from 'react'

function WalletConnect({ wallet, onConnect }) {
  const [provider, setProvider] = useState('')

  const connect = async (type) => {
    try {
      let response
      if (type === 'Phantom' && window.solana?.isPhantom) {
        response = await window.solana.connect()
      } else if (type === 'Solflare' && window.solflare?.isSolflare) {
        response = await window.solflare.connect()
      } else if (type === 'Backpack' && window.backpack?.solana) {
        response = await window.backpack.solana.connect()
      }
      if (response?.publicKey) {
        onConnect(response.publicKey.toString())
        setProvider(type)
      }
    } catch (error) {
      console.warn(error)
    }
  }

  return (
    <div className="component-card" style={{ padding: '18px' }}>
      <div className="section-label">Wallet connect</div>
      {wallet ? (
        <div>
          <p className="glow-text">Connected: {wallet.slice(0, 8)}...{wallet.slice(-8)}</p>
          <p>Provider: {provider || 'wallet'} </p>
        </div>
      ) : (
        <p>Connect a compatible Solana wallet to enable swap simulation and history.</p>
      )}
      <div className="token-list">
        {['Phantom', 'Solflare', 'Backpack'].map((item) => (
          <button key={item} className="button" type="button" onClick={() => connect(item)}>
            {wallet ? `Reconnect ${item}` : `Connect ${item}`}
          </button>
        ))}
      </div>
    </div>
  )
}

export default WalletConnect
