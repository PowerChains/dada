import { useEffect, useState } from 'react'

export function useArbitrage(apiBase) {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    setLoading(true)
    fetch(`${apiBase}/api/v1/arbitrage`)
      .then(resp => resp.json())
      .then(json => setData(json || []))
      .catch(err => setError(err.message || 'fetch error'))
      .finally(() => setLoading(false))
  }, [apiBase])

  return { data, loading, error }
}
