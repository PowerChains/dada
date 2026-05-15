import { useEffect, useState } from 'react'

export function useBestPrice(apiBase, tokenIn, tokenOut, amount, slippageTolerance) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (!tokenIn || !tokenOut) {
      return
    }
    setLoading(true)
    setError(null)
    fetch(`${apiBase}/api/v1/best-price?tokenIn=${encodeURIComponent(tokenIn)}&tokenOut=${encodeURIComponent(tokenOut)}&amount=${amount}&slippageTolerance=${slippageTolerance}`)
      .then(resp => resp.json())
      .then(json => {
        setData(json)
      })
      .catch(err => setError(err.message || 'fetch error'))
      .finally(() => setLoading(false))
  }, [apiBase, tokenIn, tokenOut, amount, slippageTolerance])

  return { data, loading, error }
}
