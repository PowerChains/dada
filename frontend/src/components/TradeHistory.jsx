import { useEffect, useMemo, useState } from 'react'

function TradeHistory({ wallet, apiBase }) {
  const [history, setHistory] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!wallet) {
      setHistory([])
      return
    }
    setLoading(true)
    fetch(`${apiBase}/api/v1/trade-history?wallet=${encodeURIComponent(wallet)}`)
      .then((res) => res.json())
      .then((json) => setHistory(json || []))
      .catch((err) => setError(err.message || 'fetch error'))
      .finally(() => setLoading(false))
  }, [wallet, apiBase])

  const analytics = useMemo(() => {
    const totalVolume = history.reduce((sum, item) => sum + (item.amount || 0), 0)
    const avgSlippage = history.length ? history.reduce((sum, item) => sum + (item.slippage || 0), 0) / history.length : 0
    const winCount = history.filter((item) => (item.slippage || 0) < 1.0).length
    const ratio = history.length ? `${((winCount / history.length) * 100).toFixed(0)}%` : '0%'
    return { totalVolume, avgSlippage, ratio }
  }, [history])

  return (
    <section className="component-card history-panel">
      <div className="section-label">Trade history</div>
      <h2>Wallet analytics</h2>
      {wallet ? (
        <>
          <div className="metric-grid">
            <div className="metric-card">
              <p className="section-label">Total volume</p>
              <p>${analytics.totalVolume.toFixed(2)}</p>
            </div>
            <div className="metric-card">
              <p className="section-label">Avg slippage</p>
              <p>{analytics.avgSlippage.toFixed(2)}%</p>
            </div>
            <div className="metric-card">
              <p className="section-label">Win rate</p>
              <p>{analytics.ratio}</p>
            </div>
          </div>
          <div className="table-wrapper" style={{ marginTop: '18px' }}>
            <table>
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Pair</th>
                  <th>Amount</th>
                  <th>Slippage</th>
                  <th>Fees</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan="5">Loading trades…</td></tr>
                ) : error ? (
                  <tr><td colSpan="5">{error}</td></tr>
                ) : history.length === 0 ? (
                  <tr><td colSpan="5">No trades yet.</td></tr>
                ) : (
                  history.map((trade) => (
                    <tr key={trade.id}>
                      <td>{new Date(trade.timestamp).toLocaleString()}</td>
                      <td>{trade.tokenIn}/{trade.tokenOut}</td>
                      <td>{trade.amount.toFixed(2)}</td>
                      <td>{trade.slippage.toFixed(2)}%</td>
                      <td>{trade.fees.toFixed(4)}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </>
      ) : (
        <p>Connect a wallet to load your trade history and analytics.</p>
      )}
    </section>
  )
}

export default TradeHistory
