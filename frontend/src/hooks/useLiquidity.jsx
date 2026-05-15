import { useEffect, useState } from 'react'

export function useLiquidity(apiBase, pair) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!pair) {
      return
    }
    setLoading(true)
    setError(null)
    fetch(`${apiBase}/api/v1/liquidity?pair=${encodeURIComponent(pair)}`)
      .then(resp => resp.json())
      .then(json => setData(json))
      .catch(err => setError(err.message || 'fetch error'))
      .finally(() => setLoading(false))
  }, [apiBase, pair])

  return { data, loading, error }
}
