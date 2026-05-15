package main

type User struct {
	Wallet            string  `json:"wallet"`
	Prefs             string  `json:"prefs"`
	SlippageTolerance float64 `json:"slippageTolerance"`
}

type Trade struct {
	ID        int64          `json:"id"`
	Wallet    string         `json:"wallet"`
	TokenIn   string         `json:"tokenIn"`
	TokenOut  string         `json:"tokenOut"`
	Amount    float64        `json:"amount"`
	Route     []RouteSegment `json:"route"`
	Fees      float64        `json:"fees"`
	Slippage  float64        `json:"slippage"`
	Timestamp string         `json:"timestamp"`
}

type PoolData struct {
	PoolID    string  `json:"poolId"`
	DEX       string  `json:"dex"`
	TokenPair string  `json:"tokenPair"`
	Liquidity float64 `json:"liquidity"`
	Fee       float64 `json:"fee"`
	Slippage  float64 `json:"slippage"`
}

type RouteSegment struct {
	DEX       string  `json:"dex"`
	PoolID    string  `json:"poolId"`
	TokenIn   string  `json:"tokenIn"`
	TokenOut  string  `json:"tokenOut"`
	AmountIn  float64 `json:"amountIn"`
	AmountOut float64 `json:"amountOut"`
	Fee       float64 `json:"fee"`
	Slippage  float64 `json:"slippage"`
}

type BestPriceResponse struct {
	TokenIn      string         `json:"tokenIn"`
	TokenOut     string         `json:"tokenOut"`
	AmountIn     float64        `json:"amountIn"`
	EstimatedOut float64        `json:"estimatedOut"`
	Slippage     float64        `json:"slippage"`
	TotalFee     float64        `json:"totalFee"`
	Route        []RouteSegment `json:"route"`
	DEXs         []string       `json:"dexs"`
	CacheKey     string         `json:"cacheKey,omitempty"`
}

type LiquidityResponse struct {
	Pair        string     `json:"pair"`
	Pools       []PoolData `json:"pools"`
	TotalTVL    float64    `json:"totalTvl"`
	AvgFee      float64    `json:"avgFee"`
	AvgSlippage float64    `json:"avgSlippage"`
}

type SimulationRequest struct {
	Wallet            string  `json:"wallet,omitempty"`
	TokenIn           string  `json:"tokenIn"`
	TokenOut          string  `json:"tokenOut"`
	Amount            float64 `json:"amount"`
	SlippageTolerance float64 `json:"slippageTolerance"`
}

type SimulationResult struct {
	Route             []RouteSegment `json:"route"`
	EstimatedOut      float64        `json:"estimatedOut"`
	TotalFee          float64        `json:"totalFee"`
	EffectiveSlippage float64        `json:"effectiveSlippage"`
	Profitability     string         `json:"profitability"`
}

type ArbitrageOpportunity struct {
	ID              int64   `json:"id,omitempty"`
	Pair            string  `json:"pair"`
	DexA            string  `json:"dexA"`
	DexB            string  `json:"dexB"`
	PriceDiff       float64 `json:"priceDiff"`
	RequiredCapital float64 `json:"requiredCapital"`
	EstProfit       float64 `json:"estProfit"`
	Timestamp       string  `json:"timestamp"`
}

type AIAssistantRequest struct {
	Message string `json:"message"`
}

type AIAssistantResponse struct {
	Suggestion string `json:"suggestion"`
}
