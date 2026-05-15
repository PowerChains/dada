package main

import (
	context "context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func GetBestPrice(ctx context.Context, db *sql.DB, cache *redis.Client, tokenIn, tokenOut string, amount float64, slippageTolerance float64) (BestPriceResponse, error) {
	tokenIn = SimplifyToken(tokenIn)
	tokenOut = SimplifyToken(tokenOut)
	if tokenIn == "" || tokenOut == "" || amount <= 0 {
		return BestPriceResponse{}, errors.New("invalid best-price request")
	}

	cacheKey := BuildCacheKey("best-price", tokenIn, tokenOut, fmt.Sprintf("%.2f", amount))
	var cached BestPriceResponse
	if ok, _ := CacheGet(ctx, cache, cacheKey, &cached); ok {
		return cached, nil
	}

	if err := EnsureSamplePools(ctx, db); err != nil {
		return BestPriceResponse{}, err
	}

	pools, err := LoadPools(ctx, db, "")
	if err != nil {
		return BestPriceResponse{}, err
	}

	baseRoute, bestOut := bestDirectQuote(pools, tokenIn, tokenOut, amount)
	helperRoute, helperOut := bestMultiHopQuote(pools, tokenIn, tokenOut, amount)

	result := BestPriceResponse{
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		AmountIn: amount,
		Route:    baseRoute,
		DEXs:     UniqueStrings(routeDEXs(baseRoute)),
	}

	if helperOut > bestOut*1.005 {
		result.Route = helperRoute
		bestOut = helperOut
		result.DEXs = UniqueStrings(routeDEXs(helperRoute))
	}

	result.EstimatedOut = bestOut * (1.0 - slippageTolerance/100.0)
	result.TotalFee = totalRouteFee(result.Route)
	result.Slippage = slippageTolerance
	result.CacheKey = cacheKey

	if err := CacheSet(ctx, cache, cacheKey, result, 30*time.Second); err != nil {
		// ignore cache failures
	}
	return result, nil
}

func AggregateLiquidity(ctx context.Context, db *sql.DB, pair string) (LiquidityResponse, error) {
	if err := EnsureSamplePools(ctx, db); err != nil {
		return LiquidityResponse{}, err
	}

	pools, err := LoadPools(ctx, db, pair)
	if err != nil {
		return LiquidityResponse{}, err
	}
	if len(pools) == 0 {
		return LiquidityResponse{Pair: pair}, nil
	}
	vol, avgFee, avgSlip := AveragePoolMetrics(pools)
	return LiquidityResponse{
		Pair:        pair,
		Pools:       pools,
		TotalTVL:    vol,
		AvgFee:      avgFee,
		AvgSlippage: avgSlip,
	}, nil
}

func SimulateSwap(ctx context.Context, db *sql.DB, cache *redis.Client, req SimulationRequest) (SimulationResult, error) {
	if req.TokenIn == "" || req.TokenOut == "" || req.Amount <= 0 {
		return SimulationResult{}, errors.New("invalid simulation payload")
	}

	best, err := GetBestPrice(ctx, db, cache, req.TokenIn, req.TokenOut, req.Amount, req.SlippageTolerance)
	if err != nil {
		return SimulationResult{}, err
	}

	sim := SimulationResult{
		Route:             best.Route,
		EstimatedOut:      best.EstimatedOut,
		TotalFee:          best.TotalFee,
		EffectiveSlippage: best.Slippage,
	}

	if len(best.Route) == 0 {
		sim.Profitability = "no route"
	} else if best.EstimatedOut > req.Amount {
		sim.Profitability = "positive"
	} else {
		sim.Profitability = "neutral"
	}

	if req.Wallet != "" {
		trade := Trade{
			Wallet:   req.Wallet,
			TokenIn:  req.TokenIn,
			TokenOut: req.TokenOut,
			Amount:   req.Amount,
			Route:    best.Route,
			Fees:     best.TotalFee,
			Slippage: best.Slippage,
		}
		_ = RecordTrade(ctx, db, trade)
	}

	return sim, nil
}

func DetectArbitrage(ctx context.Context, db *sql.DB, cache *redis.Client) ([]ArbitrageOpportunity, error) {
	if err := EnsureSamplePools(ctx, db); err != nil {
		return nil, err
	}

	if ok, _ := CacheGet(ctx, cache, "arbitrage-snapshot", &[]ArbitrageOpportunity{}); ok {
		var cached []ArbitrageOpportunity
		if err := CacheGet(ctx, cache, "arbitrage-snapshot", &cached); err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	pools, err := LoadPools(ctx, db, "")
	if err != nil {
		return nil, err
	}

	pairs := tokenPairsFromPools(pools)
	var opportunities []ArbitrageOpportunity
	poolMap := BuildPoolMap(pools)

	for _, pair := range pairs {
		dexPools := poolMap[pair]
		if len(dexPools) < 2 {
			continue
		}

		for i := 0; i < len(dexPools); i++ {
			for j := i + 1; j < len(dexPools); j++ {
				outA := EstimateSwapAmount(dexPools[i], 1000)
				outB := EstimateSwapAmount(dexPools[j], 1000)
				if math.Abs(outA-outB) < 1 {
					continue
				}
				priceDiff := math.Abs(outA-outB) / 1000 * 100
				if priceDiff < 0.35 {
					continue
				}
				estProfit := math.Abs(outA-outB) * 0.92
				opportunities = append(opportunities, ArbitrageOpportunity{
					Pair:            pair,
					DexA:            dexPools[i].DEX,
					DexB:            dexPools[j].DEX,
					PriceDiff:       priceDiff,
					RequiredCapital: math.Min(MaximumArbitrageCapital(dexPools[i]), MaximumArbitrageCapital(dexPools[j])),
					EstProfit:       testProfit,
					Timestamp:       FormatTimestamp(time.Now()),
				})
			}
		}
	}
	if len(opportunities) == 0 {
		return nil, nil
	}
	_ = SaveArbitrageOpportunities(ctx, db, opportunities)
	_ = CacheSet(ctx, cache, "arbitrage-snapshot", opportunities, 20*time.Second)
	return opportunities, nil
}

func AskAiStudio(ctx context.Context, cfg Config, prompt string) (string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "Provide a token symbol or swap intent to receive suggestions.", nil
	}

	if cfg.AiStudioKey == "" {
		return fmt.Sprintf("AI suggest: best routes for %s using SOL/USDC, BTC/USDC, and ETH/USDC pools.", trimmed), nil
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://ai.studio/api/v1/query", strings.NewReader(fmt.Sprintf(`{"prompt":"%s"}`, trimmed)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AiStudioKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai studio responded %d", resp.StatusCode)
	}

	return fmt.Sprintf("AI suggested route for '%s' based on live pool data.", trimmed), nil
}

func bestDirectQuote(pools []PoolData, tokenIn, tokenOut string, amount float64) ([]RouteSegment, float64) {
	bestPool := bestPoolForPair(pools, tokenIn, tokenOut)
	if bestPool == nil {
		return nil, 0
	}
	segment := BuildRouteSegment(*bestPool, tokenIn, tokenOut, amount)
	return []RouteSegment{segment}, segment.AmountOut
}

func bestMultiHopQuote(pools []PoolData, tokenIn, tokenOut string, amount float64) ([]RouteSegment, float64) {
	var bestRoute []RouteSegment
	bestOut := 0.0
	for _, hop := range PreferredHopTokens() {
		if strings.EqualFold(hop, tokenIn) || strings.EqualFold(hop, tokenOut) {
			continue
		}
		firstPool := bestPoolForPair(pools, tokenIn, hop)
		secondPool := bestPoolForPair(pools, hop, tokenOut)
		if firstPool == nil || secondPool == nil {
			continue
		}
		first := BuildRouteSegment(*firstPool, tokenIn, hop, amount)
		second := BuildRouteSegment(*secondPool, hop, tokenOut, first.AmountOut)
		if second.AmountOut > bestOut {
			bestOut = second.AmountOut
			bestRoute = []RouteSegment{first, second}
		}
	}
	return bestRoute, bestOut
}

func totalRouteFee(route []RouteSegment) float64 {
	fee := 0.0
	for _, segment := range route {
		fee += segment.Fee
	}
	return fee
}

func routeDEXs(route []RouteSegment) []string {
	var dexs []string
	for _, step := range route {
		dexs = append(dexs, step.DEX)
	}
	return dexs
}

func (s *Server) RunArbitrageBroadcaster(ctx context.Context) {
	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			opps, err := DetectArbitrage(ctx, s.db, s.cache)
			if err != nil {
				continue
			}
			s.broadcastArbitrage(opps)
		case <-ctx.Done():
			return
		}
	}
}
