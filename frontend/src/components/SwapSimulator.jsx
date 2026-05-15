import { useMemo, useState } from 'react'
import { useSwapSimulation } from '../hooks/useSwapSimulation.jsx'

function SwapSimulator({ wallet, selectedPair, slippageTolerance, onSlippageChange, apiBase }) {
  const [amount, setAmount] = useState(950)
  const [estimated, setEstimated] = useState(null)
  const [routeVisual, setRouteVisual] = useState([])
  const [status, setStatus] = useState('')

  const { data, loading, error } = useSwapSimulation(apiBase, {
    wallet,
    tokenIn: selectedPair.split('/')[0],
    tokenOut: selectedPair.split('/')[1],
    amount,
    slippageTolerance,
  })

  const handleSimulate = async () => {
    if (!wallet) {
      setStatus('Connect a wallet to save trade history and simulate with your preferences.')
      return
    }
    setStatus('Simulation performed for wallet: ' + wallet.slice(0, 8))
    setEstimated(data)
    setRouteVisual(data?.route || [])
  }

  const routeDetails = useMemo(() => {
    return routeVisual.map((step, index) => (
      <div key={index} className="route-step">
        <span>{step.tokenIn} → {step.tokenOut}</span>
        <span>{step.dex}</span>
      </div>
    ))
  }, [routeVisual])

  return (
    <section className="component-card swap-panel">
      <div className="section-label">Swap simulator</div>
      <h2>Swap preview</h2>
      <div className="form-row">
        <label>
          Amount to swap
          <input className="input" type="number" value={amount} onChange={(e) => setAmount(Number(e.target.value))} />
        </label>
        <label>
          Slippage tolerance
          <input className="input" type="range" min="0" max="5" step="0.1" value={slippageTolerance} onChange={(e) => onSlippageChange(Number(e.target.value))} />
          <div>{slippageTolerance.toFixed(1)}%</div>
        </label>
      </div>
      <button className="button" type="button" onClick={handleSimulate}>Run simulation</button>
      <div className="metric-grid" style={{ marginTop: '18px' }}>
        <div className="metric-card">
          <p className="section-label">Estimated output</p>
          <p className="glow-text">{loading ? '...' : data ? data.estimatedOut.toFixed(4) : '--'}</p>
        </div>
        <div className="metric-card">
          <p className="section-label">Route impact</p>
          <p>{data ? `${data.effectiveSlippage.toFixed(2)}%` : '--'}</p>
        </div>
      </div>
      {error && <p style={{ color: '#ff8da8' }}>{error}</p>}
      <div className="route-visual">
        {routeDetails.length ? routeDetails : <div className="route-step">No route yet.</div>}
      </div>
      {status && <p style={{ marginTop: '14px', color: '#b8c7ff' }}>{status}</p>}
    </section>
  )
}

export default SwapSimulator
