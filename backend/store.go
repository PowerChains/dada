package main

import (
	context "context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func initSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS users (
    wallet TEXT PRIMARY KEY,
    prefs JSONB DEFAULT '{}'::jsonb,
    slippage_tolerance DOUBLE PRECISION DEFAULT 0.5
);

CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,
    wallet TEXT NOT NULL,
    token_in TEXT NOT NULL,
    token_out TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    route JSONB NOT NULL,
    fees DOUBLE PRECISION NOT NULL,
    slippage DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dex_data (
    pool_id TEXT PRIMARY KEY,
    dex TEXT NOT NULL,
    token_pair TEXT NOT NULL,
    liquidity DOUBLE PRECISION NOT NULL,
    fee DOUBLE PRECISION NOT NULL,
    slippage DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS arbitrage_opportunities (
    id SERIAL PRIMARY KEY,
    pair TEXT NOT NULL,
    dex_a TEXT NOT NULL,
    dex_b TEXT NOT NULL,
    price_diff DOUBLE PRECISION NOT NULL,
    required_capital DOUBLE PRECISION NOT NULL,
    est_profit DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}

func LoadTradeHistory(ctx context.Context, db *sql.DB, wallet string) ([]Trade, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, wallet, token_in, token_out, amount, route, fees, slippage, timestamp FROM trades WHERE wallet = $1 ORDER BY timestamp DESC LIMIT 50`, wallet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []Trade
	for rows.Next() {
		var t Trade
		var routeJson []byte
		if err := rows.Scan(&t.ID, &t.Wallet, &t.TokenIn, &t.TokenOut, &t.Amount, &routeJson, &t.Fees, &t.Slippage, &t.Timestamp); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(routeJson, &t.Route); err != nil {
			return nil, err
		}
		history = append(history, t)
	}
	return history, nil
}

func RecordTrade(ctx context.Context, db *sql.DB, trade Trade) error {
	routeData, err := json.Marshal(trade.Route)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO trades (wallet, token_in, token_out, amount, route, fees, slippage) VALUES ($1, $2, $3, $4, $5, $6, $7)`, trade.Wallet, trade.TokenIn, trade.TokenOut, trade.Amount, routeData, trade.Fees, trade.Slippage)
	return err
}

func LoadPools(ctx context.Context, db *sql.DB, pairFilter string) ([]PoolData, error) {
	query := `SELECT pool_id, dex, token_pair, liquidity, fee, slippage FROM dex_data`
	args := []interface{}{}
	if pairFilter != "" {
		query += ` WHERE token_pair ILIKE '%' || $1 || '%'`
		args = append(args, pairFilter)
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []PoolData
	for rows.Next() {
		var p PoolData
		if err := rows.Scan(&p.PoolID, &p.DEX, &p.TokenPair, &p.Liquidity, &p.Fee, &p.Slippage); err != nil {
			return nil, err
		}
		pools = append(pools, p)
	}
	return pools, nil
}

func SaveDexData(ctx context.Context, db *sql.DB, pools []PoolData) error {
	stmt, err := db.PrepareContext(ctx, `INSERT INTO dex_data (pool_id, dex, token_pair, liquidity, fee, slippage) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (pool_id) DO UPDATE SET dex = EXCLUDED.dex, token_pair = EXCLUDED.token_pair, liquidity = EXCLUDED.liquidity, fee = EXCLUDED.fee, slippage = EXCLUDED.slippage`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, pool := range pools {
		if _, err := stmt.ExecContext(ctx, pool.PoolID, pool.DEX, pool.TokenPair, pool.Liquidity, pool.Fee, pool.Slippage); err != nil {
			return err
		}
	}
	return nil
}

func SaveArbitrageOpportunities(ctx context.Context, db *sql.DB, opportunities []ArbitrageOpportunity) error {
	stmt, err := db.PrepareContext(ctx, `INSERT INTO arbitrage_opportunities (pair, dex_a, dex_b, price_diff, required_capital, est_profit) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, opp := range opportunities {
		if _, err := stmt.ExecContext(ctx, opp.Pair, opp.DexA, opp.DexB, opp.PriceDiff, opp.RequiredCapital, opp.EstProfit); err != nil {
			return err
		}
	}
	return nil
}

func LoadRecentArbitrage(ctx context.Context, db *sql.DB, limit int) ([]ArbitrageOpportunity, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, pair, dex_a, dex_b, price_diff, required_capital, est_profit, timestamp FROM arbitrage_opportunities ORDER BY timestamp DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []ArbitrageOpportunity
	for rows.Next() {
		var item ArbitrageOpportunity
		if err := rows.Scan(&item.ID, &item.Pair, &item.DexA, &item.DexB, &item.PriceDiff, &item.RequiredCapital, &item.EstProfit, &item.Timestamp); err != nil {
			return nil, err
		}
		opps = append(opps, item)
	}
	return ops, nil
}

func CacheSet(ctx context.Context, cache *redis.Client, key string, value interface{}, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cache.Set(ctx, key, payload, ttl).Err()
}

func CacheGet(ctx context.Context, cache *redis.Client, key string, dest interface{}) (bool, error) {
	payload, err := cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(payload, dest); err != nil {
		return false, err
	}
	return true, nil
}

func NormalizePair(tokenIn, tokenOut string) string {
	return strings.ToUpper(strings.TrimSpace(tokenIn)) + "/" + strings.ToUpper(strings.TrimSpace(tokenOut))
}

func SplitPair(pair string) (string, string) {
	parts := strings.Split(pair, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func UniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func TokenListForPair(pair string) []string {
	parts := strings.Split(pair, "/")
	if len(parts) != 2 {
		return []string{pair}
	}
	return []string{parts[0], parts[1]}
}

func PoolTokenPair(pool PoolData) (string, string) {
	parts := strings.Split(pool.TokenPair, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func MatchingPools(pools []PoolData, tokenIn, tokenOut string) []PoolData {
	want1 := NormalizePair(tokenIn, tokenOut)
	want2 := NormalizePair(tokenOut, tokenIn)
	var matches []PoolData
	for _, pool := range pools {
		if NormalizePair(strings.Split(pool.TokenPair, "/")[0], strings.Split(pool.TokenPair, "/")[1]) == want1 || NormalizePair(strings.Split(pool.TokenPair, "/")[0], strings.Split(pool.TokenPair, "/")[1]) == want2 {
			matches = append(matches, pool)
		}
	}
	return matches
}

func AveragePoolMetrics(pools []PoolData) (float64, float64, float64) {
	if len(pools) == 0 {
		return 0, 0, 0
	}
	tvl := 0.0
	fee := 0.0
	slip := 0.0
	for _, p := range pools {
		tvl += p.Liquidity
		fee += p.Fee
		slip += p.Slippage
	}
	return tvl, fee / float64(len(pools)), slip / float64(len(pools))
}

func EnsureSamplePools(ctx context.Context, db *sql.DB) error {
	pools := generateSamplePools()
	return SaveDexData(ctx, db, pools)
}

func generateSamplePools() []PoolData {
	return []PoolData{
		{"pool-meteora-usdc-usdt", "Meteora", "USDC/USDT", 12_500_000, 0.0004, 0.35},
		{"pool-raydium-sol-usdc", "Raydium", "SOL/USDC", 18_200_000, 0.00045, 0.4},
		{"pool-orca-btc-usdc", "Orca", "BTC/USDC", 9_800_000, 0.00025, 0.35},
		{"pool-kamino-eth-usdc", "Kamino", "ETH/USDC", 7_400_000, 0.00035, 0.5},
		{"pool-jupiter-sol-usdc", "Jupiter", "SOL/USDC", 22_100_000, 0.0003, 0.38},
		{"pool-aldrin-eth-sol", "Aldrin", "ETH/SOL", 6_300_000, 0.0004, 0.48},
		{"pool-duality-raydium-usdc", "Duality", "RAY/USDC", 5_200_000, 0.0004, 0.45},
		{"pool-openbook-btc-sol", "OpenBook", "BTC/SOL", 4_100_000, 0.0003, 0.4},
		{"pool-serum-usdc-usdt", "Serum", "USDC/USDT", 3_900_000, 0.00018, 0.35},
		{"pool-jet-sol-usdt", "Jet", "SOL/USDT", 8_500_000, 0.00028, 0.42},
		{"pool-saber-eth-usdt", "Saber", "ETH/USDT", 5_700_000, 0.00033, 0.46},
		{"pool-sunrise-btc-usdt", "Sunrise", "BTC/USDT", 4_500_000, 0.00027, 0.4},
		{"pool-atrix-raydium-usdt", "Atrix", "RAY/USDT", 2_700_000, 0.00038, 0.47},
		{"pool-meteor-kamino-sol-btc", "Meteora", "SOL/BTC", 2_100_000, 0.00032, 0.41},
		{"pool-wormhole-eth-btc", "Wormhole", "ETH/BTC", 3_200_000, 0.00029, 0.43},
		{"pool-lifinity-usdc-btc", "Lifinity", "USDC/BTC", 2_900_000, 0.00035, 0.44},
		{"pool-triton-sol-eth", "Triton", "SOL/ETH", 4_000_000, 0.00034, 0.39},
		{"pool-mercurial-btc-usdc", "Mercurial", "BTC/USDC", 6_100_000, 0.00026, 0.36},
		{"pool-atlas-sol-usdc", "Sunrise", "SOL/USDC", 12_900_000, 0.00031, 0.37},
		{"pool-poolx-raydium-sol", "PoolX", "RAY/SOL", 1_900_000, 0.00037, 0.44},
		{"pool-duality-usdc-magic", "Duality", "USDC/MAGIC", 1_200_000, 0.0005, 0.6},
		{"pool-openbook-civic-usdc", "OpenBook", "CIVIC/USDC", 950_000, 0.0005, 0.55},
		{"pool-jet-eth-btc", "Jet", "ETH/BTC", 3_600_000, 0.00028, 0.42},
		{"pool-serum-sol-btc", "Serum", "SOL/BTC", 5_500_000, 0.00029, 0.41},
	}
}

func MergePopularDEXs() []string {
	return []string{"Meteora", "Raydium", "Orca", "Kamino", "Jupiter", "Aldrin", "Duality", "OpenBook", "Serum", "Jet", "Saber", "Sunrise", "Atrix", "Wormhole", "Lifinity", "Triton", "Mercurial", "PoolX", "Drift", "Mercurial", "Sunrise", "Jet", "Triton"}
}

func ParseRoutePair(route []RouteSegment) (string, string) {
	if len(route) == 0 {
		return "", ""
	}
	return route[0].TokenIn, route[len(route)-1].TokenOut
}

func CopyRoute(route []RouteSegment) []RouteSegment {
	copyRoute := make([]RouteSegment, len(route))
	copy(copyRoute, route)
	return copyRoute
}

func TokenAddress(token string) string {
	return strings.ToUpper(token)
}

func BuildSearchSuggestions(input string) []string {
	symbols := []string{"USDC", "USDT", "SOL", "BTC", "ETH", "RAY", "MANA", "CIVIC", "COPE", "GRAPE", "MANGO", "OXY", "MAGIC"}
	matches := []string{}
	for _, symbol := range symbols {
		if len(input) == 0 || strings.HasPrefix(strings.ToUpper(symbol), strings.ToUpper(input)) {
			matches = append(matches, symbol)
		}
	}
	return matches
}

func Multiply(a, b float64) float64 {
	return a * b
}

func IdentifyTokens(pair string) []string {
	parts := strings.Split(pair, "/")
	return parts
}

func TrimToken(token string) string {
	return strings.TrimSpace(strings.ToUpper(token))
}

func BuildCacheKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func FormatPrice(amount float64) float64 {
	return amount
}

func RateImpact(liquidity, amount float64) float64 {
	if liquidity <= 0 {
		return 0.0
	}
	impact := (amount / (liquidity + amount)) * 100
	if impact > 8.0 {
		impact = 8.0
	}
	return impact
}

func DecomposePair(pair string) (string, string) {
	parts := strings.Split(pair, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func PreferredHopTokens() []string {
	return []string{"USDC", "USDT", "SOL", "BTC", "ETH", "RAY"}
}

func PriceImpactForPool(pool PoolData, amount float64) float64 {
	impact := RateImpact(pool.Liquidity, amount)
	return impact + pool.Slippage
}

func BuildRouteSegment(pool PoolData, tokenIn, tokenOut, amountIn float64) RouteSegment {
	amountOut := EstimateSwapAmount(pool, amountIn)
	return RouteSegment{
		DEX:       pool.DEX,
		PoolID:    pool.PoolID,
		TokenIn:   tokenIn,
		TokenOut:  tokenOut,
		AmountIn:  amountIn,
		AmountOut: amountOut,
		Fee:       pool.Fee * amountIn,
		Slippage:  pool.Slippage,
	}
}

func EstimateSwapAmount(pool PoolData, amountIn float64) float64 {
	priceFactor := 1.0 - pool.Fee - pool.Slippage/100.0
	impact := RateImpact(pool.Liquidity, amountIn) / 100.0
	if impact > 0.05 {
		priceFactor -= impact
	}
	if priceFactor < 0.75 {
		priceFactor = 0.75
	}
	return amountIn * priceFactor
}

func EstimatePrice(tokenIn, tokenOut string, amountIn float64, pools []PoolData) float64 {
	best := 0.0
	for _, pool := range pools {
		if MatchTokenPair(pool.TokenPair, tokenIn, tokenOut) {
			estimate := EstimateSwapAmount(pool, amountIn)
			if estimate > best {
				best = estimate
			}
		}
	}
	return best
}

func MatchTokenPair(pair, tokenIn, tokenOut string) bool {
	parts := strings.Split(pair, "/")
	if len(parts) != 2 {
		return false
	}
	return (strings.EqualFold(parts[0], tokenIn) && strings.EqualFold(parts[1], tokenOut)) || (strings.EqualFold(parts[0], tokenOut) && strings.EqualFold(parts[1], tokenIn))
}

func BuildDefaultPools() []PoolData {
	return generateSamplePools()
}

func PoolKey(pool PoolData) string {
	return pool.DEX + ":" + pool.PoolID
}

func SimplifyToken(token string) string {
	return strings.ToUpper(strings.TrimSpace(token))
}

func PoolPairsByToken(pools []PoolData, token string) []PoolData {
	var found []PoolData
	for _, pool := range pools {
		tokens := strings.Split(pool.TokenPair, "/")
		if len(tokens) != 2 {
			continue
		}
		if strings.EqualFold(tokens[0], token) || strings.EqualFold(tokens[1], token) {
			found = append(found, pool)
		}
	}
	return found
}

func PoolMarketDepth(pool PoolData) float64 {
	return pool.Liquidity * (1.0 - pool.Fee)
}

func AveragePoolPrice(pool PoolData, amountIn float64) float64 {
	return EstimateSwapAmount(pool, amountIn)
}

func MaximumArbitrageCapital(pool PoolData) float64 {
	limit := pool.Liquidity * 0.08
	if limit < 1000 {
		return 1000
	}
	return limit
}

func EffectiveSpread(poolA, poolB PoolData) float64 {
	return poolA.Slippage + poolB.Slippage + poolA.Fee + poolB.Fee
}

func ComparePools(poolA, poolB PoolData, amount float64) (float64, float64) {
	outA := EstimateSwapAmount(poolA, amount)
	outB := EstimateSwapAmount(poolB, amount)
	return outA, outB
}

func GenerateQuickPairs() []string {
	return []string{"SOL/USDC", "BTC/USDC", "ETH/USDC", "USDC/USDT", "RAY/USDC", "SOL/ETH"}
}

func BuildArbitrageQuote(poolA, poolB PoolData, amount float64) (float64, float64) {
	outA := EstimateSwapAmount(poolA, amount)
	outB := EstimateSwapAmount(poolB, amount)
	profit := outB - outA
	return outA, profit
}

func tokenPairsFromPools(pools []PoolData) []string {
	seen := map[string]struct{}{}
	pairs := []string{}
	for _, pool := range pools {
		pair := NormalizePair(strings.Split(pool.TokenPair, "/")[0], strings.Split(pool.TokenPair, "/")[1])
		if _, ok := seen[pair]; ok {
			continue
		}
		seen[pair] = struct{}{}
		pairs = append(pairs, pair)
	}
	return pairs
}

func BuildPoolMap(pools []PoolData) map[string][]PoolData {
	m := map[string][]PoolData{}
	for _, pool := range pools {
		pair := NormalizePair(strings.Split(pool.TokenPair, "/")[0], strings.Split(pool.TokenPair, "/")[1])
		m[pair] = append(m[pair], pool)
	}
	return m
}

func TotalPoolLiquidity(pools []PoolData) float64 {
	total := 0.0
	for _, p := range pools {
		total += p.Liquidity
	}
	return total
}

func BuildTokenGraph(pools []PoolData) map[string][]PoolData {
	graph := map[string][]PoolData{}
	for _, pool := range pools {
		parts := strings.Split(pool.TokenPair, "/")
		if len(parts) != 2 {
			continue
		}
		graph[strings.ToUpper(parts[0])] = append(graph[strings.ToUpper(parts[0])], pool)
		graph[strings.ToUpper(parts[1])] = append(graph[strings.ToUpper(parts[1])], pool)
	}
	return graph
}

func bestPoolForPair(pools []PoolData, tokenIn, tokenOut string) *PoolData {
	var best *PoolData
	bestScore := -1.0
	for _, pool := range pools {
		if MatchTokenPair(pool.TokenPair, tokenIn, tokenOut) {
			score := pool.Liquidity - pool.Fee*1_000_000 - pool.Slippage*100
			if score > bestScore {
				bestScore = score
				best = &pool
			}
		}
	}
	return best
}
