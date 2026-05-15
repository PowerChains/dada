import { useEffect, useState } from 'react'

export function useSwapSimulation(apiBase, request) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!request?.tokenIn || !request?.tokenOut || !request?.amount) {
      return
    }
    setLoading(true)
    setError(null)
    fetch(`${apiBase}/api/v1/simulate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    })
      .then(resp => resp.json())
      .then(json => setData(json))
      .catch(err => setError(err.message || 'fetch error'))
      .finally(() => setLoading(false))
  }, [apiBase, request])

  return { data, loading, error }
}
