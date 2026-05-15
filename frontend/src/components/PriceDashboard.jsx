import { useMemo, useState } from 'react'
import { useBestPrice } from '../hooks/useBestPrice.jsx'
import { useLiquidity } from '../hooks/useLiquidity.jsx'

const pairOptions = ['SOL/USDC', 'BTC/USDC', 'ETH/USDC', 'USDC/USDT', 'RAY/USDC']

function PriceDashboard({ selectedPair, onPairChange, slippageTolerance, apiBase }) {
  const [amount, setAmount] = useState(1200)
  const [tokenIn, tokenOut] = selectedPair.split('/')
  const { data, loading } = useBestPrice(apiBase, tokenIn, tokenOut, amount, slippageTolerance)
  const { data: liquidity } = useLiquidity(apiBase, selectedPair)

  const routeSummary = useMemo(() => {
    if (!data?.route) return 'No route available.'
    return data.route.map((step) => `${step.tokenIn} → ${step.tokenOut} @ ${step.dex}`).join(' / ')
  }, [data])

  return (
    <section className="component-card price-panel">
      <div className="section-label">Price dashboard</div>
      <h2>Best-price finder</h2>
      <label className="form-row">
        Swap pair
        <select className="select" value={selectedPair} onChange={(e) => onPairChange(e.target.value)}>
          {pairOptions.map((pair) => (
            <option key={pair} value={pair}>{pair}</option>
          ))}
        </select>
      </label>
      <label className="form-row">
        Amount
        <input className="input" type="number" value={amount} onChange={(e) => setAmount(Number(e.target.value))} />
      </label>
      <div className="metric-grid">
        <div className="metric-card">
          <p className="section-label">Estimated output</p>
          <p className="glow-text">{loading ? 'Loading…' : data ? data.estimatedOut.toFixed(4) : '--'}</p>
        </div>
        <div className="metric-card">
          <p className="section-label">Fee impact</p>
          <p>{data ? `${data.totalFee.toFixed(4)} tokens` : '--'}</p>
        </div>
      </div>
      <div className="route-visual">
        <div className="route-step">
          <span>Route</span>
          <span>{routeSummary}</span>
        </div>
        <div className="route-step">
          <span>Slippage</span>
          <span>{data ? `${data.slippage.toFixed(2)}%` : '--'}</span>
        </div>
      </div>
      <div className="metric-card" style={{ marginTop: '18px' }}>
        <p className="section-label">Liquidity snapshot</p>
        <p>{liquidity ? `${liquidity.totalTvl.toFixed(0)} TVL across ${liquidity.pools?.length ?? 0} pools` : 'Loading pools...'}</p>
      </div>
    </section>
  )
}

export default PriceDashboard
