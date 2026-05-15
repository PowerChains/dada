import { useEffect, useMemo, useState } from 'react'

const popular = ['SOL', 'USDC', 'USDT', 'BTC', 'ETH', 'RAY', 'MANA']

function TokenSearch({ selectedPair, onPairChange, aiPrompt, onAiPromptChange, suggestion, onAiSuggest, apiBase }) {
  const [search, setSearch] = useState('')
  const [suggestions, setSuggestions] = useState(popular)
  const [lastResponse, setLastResponse] = useState('')

  useEffect(() => {
    const query = aiPrompt.trim()
    if (!query) {
      return
    }
    fetch(`${apiBase}/api/v1/ai`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: query }),
    })
      .then(res => res.json())
      .then(json => {
        setLastResponse(json.suggestion || json.error || '')
      })
      .catch(() => setLastResponse('AI assistant unavailable.'))
  }, [aiPrompt, apiBase])

  const filtered = useMemo(() => {
    if (!search) return popular
    return popular.filter((token) => token.toLowerCase().startsWith(search.toLowerCase()))
  }, [search])

  return (
    <section className="component-card search-panel">
      <div className="section-label">Token search</div>
      <h2>AI search + suggestions</h2>
      <label className="form-row">
        Search token
        <input className="input" value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Type SOL, USDC or BTC" />
      </label>
      <div className="token-list">
        {filtered.slice(0, 6).map((symbol) => (
          <button key={symbol} className="button" type="button" onClick={() => onPairChange(`${symbol}/${selectedPair.split('/')[1]}`)}>
            {symbol}
          </button>
        ))}
      </div>
      <label className="form-row">
        Smart suggestion
        <input className="input" value={aiPrompt} onChange={(e) => { onAiPromptChange(e.target.value); setSearch(e.target.value) }} placeholder="Ask DADA for best swap ideas" />
      </label>
      <div className="metric-card" style={{ marginTop: '18px' }}>
        <p className="section-label">AI hint</p>
        <p>{lastResponse || 'Type a pair prompt to get suggestions.'}</p>
      </div>
      <div className="token-list">
        {suggestions.map((token) => (
          <span key={token} className="token-chip">{token}</span>
        ))}
      </div>
    </section>
  )
}

export default TokenSearch
