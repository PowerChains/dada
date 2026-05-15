import { useEffect, useMemo, useState } from 'react'
import WalletConnect from './components/WalletConnect.jsx'
import PriceDashboard from './components/PriceDashboard.jsx'
import SwapSimulator from './components/SwapSimulator.jsx'
import TokenSearch from './components/TokenSearch.jsx'
import TradeHistory from './components/TradeHistory.jsx'

const API_BASE = import.meta.env.VITE_API_BASE ?? ''
const WS_BASE = API_BASE
  ? API_BASE.replace(/^http/, 'ws')
  : `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}`

function App() {
  const [wallet, setWallet] = useState('')
  const [slippageTolerance, setSlippageTolerance] = useState(0.5)
  const [selectedPair, setSelectedPair] = useState('SOL/USDC')
  const [aiPrompt, setAiPrompt] = useState('')
  const [aiSuggestion, setAiSuggestion] = useState('')
  const [arbitrageFeed, setArbitrageFeed] = useState([])

  useEffect(() => {
    const stored = window.localStorage.getItem('dada-slippage')
    if (stored) {
      setSlippageTolerance(Number(stored))
    }
    const ws = new WebSocket(`${WS_BASE}/ws/market`)
    ws.addEventListener('message', event => {
      try {
        const payload = JSON.parse(event.data)
        if (payload.type === 'arbitrage' && Array.isArray(payload.items)) {
          setArbitrageFeed(payload.items)
        }
      } catch (error) {
        console.warn(error)
      }
    })
    return () => ws.close()
  }, [])

  useEffect(() => {
    window.localStorage.setItem('dada-slippage', slippageTolerance)
  }, [slippageTolerance])

  const pairTokens = useMemo(() => selectedPair.split('/'), [selectedPair])

  return (
    <div className="app-shell">
      <div className="background-grid" />
      <header className="topbar">
        <div>
          <h1>DADA</h1>
          <p>Solana multi-DEX aggregator with predictive routing and arbitrage intelligence.</p>
        </div>
        <WalletConnect wallet={wallet} onConnect={setWallet} />
      </header>

      <main>
        <section className="hero-panel glass-card">
          <div>
            <h2>Best routes, simulated swaps, and market arbitrage.</h2>
            <p>Connect your wallet, choose a pair, and preview routes across 22+ Solana DEXs.</p>
          </div>
          <div className="arbitrage-panel">
            <h3>Live arbitrage feed</h3>
            <ul>
              {arbitrageFeed.slice(0, 4).map((item, index) => (
                <li key={index}>
                  <strong>{item.pair}</strong> {item.dexA} ↔ {item.dexB} • +{item.priceDiff.toFixed(2)}% • est. ${item.estProfit.toFixed(2)}
                </li>
              ))}
            </ul>
          </div>
        </section>

        <section className="grid-layout">
          <PriceDashboard
            selectedPair={selectedPair}
            onPairChange={setSelectedPair}
            slippageTolerance={slippageTolerance}
            apiBase={API_BASE}
          />
          <SwapSimulator
            wallet={wallet}
            selectedPair={selectedPair}
            slippageTolerance={slippageTolerance}
            onSlippageChange={setSlippageTolerance}
            apiBase={API_BASE}
          />
          <TokenSearch
            selectedPair={selectedPair}
            onPairChange={setSelectedPair}
            aiPrompt={aiPrompt}
            onAiPromptChange={setAiPrompt}
            suggestion={aiSuggestion}
            onAiSuggest={setAiSuggestion}
            apiBase={API_BASE}
          />
          <TradeHistory wallet={wallet} apiBase={API_BASE} />
        </section>
      </main>
    </div>
  )
}

export default App
