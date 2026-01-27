package services

import (
	"algebra-apr-backend/internal/client"
	"algebra-apr-backend/internal/graphql"
	"algebra-apr-backend/internal/logger"
	"algebra-apr-backend/internal/models"
	"algebra-apr-backend/internal/types"
	"algebra-apr-backend/internal/utils"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type APRService struct {
	db *gorm.DB
}

func NewAPRService(db *gorm.DB) *APRService {
	return &APRService{
		db: db,
	}
}

// Calculate all APR values in one go - optimized approach
func (s *APRService) getClientsForNetwork(networkID uint) (*client.GraphQLClient, *client.GraphQLClient, error) {
	var network models.Network
	if err := s.db.First(&network, networkID).Error; err != nil {
		return nil, nil, fmt.Errorf("network not found: %w", err)
	}

	analyticsClient := client.NewGraphQLClient(network.AnalyticsSubgraphURL, network.APIKey)
	farmingClient := client.NewGraphQLClient(network.FarmingSubgraphURL, network.APIKey)

	return analyticsClient, farmingClient, nil
}

// Calculate all APR values in one go - optimized approach
func (s *APRService) UpdateAllAPR(networkID uint) error {
	var network models.Network
	if err := s.db.First(&network, networkID).Error; err != nil {
		return fmt.Errorf("network not found: %w", err)
	}

	analyticsClient, farmingClient, err := s.getClientsForNetwork(networkID)
	if err != nil {
		return err
	}

	logger.Logger.Info("Starting full APR update", zap.String("network", network.Title))

	// Get all pools in one request
	pools, err := s.getAllPools(analyticsClient)
	if err != nil {
		return fmt.Errorf("failed to get pools: %w", err)
	}

	// Get pool day data for yesterday
	poolDayDatas, err := s.getPoolDayDatas(analyticsClient)
	if err != nil {
		return fmt.Errorf("failed to get pool day data: %w", err)
	}

	// Create map for quick lookup of pool fees
	poolFeesMap := make(map[string]types.PoolDayData)
	for _, poolDayData := range poolDayDatas {
		poolFeesMap[poolDayData.Pool.ID] = poolDayData
	}

	// Get all positions in one request
	positions, err := s.getAllPositions(analyticsClient)
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	positionsById := make(map[string]types.Position, len(positions))
	for _, position := range positions {
		positionsById[position.ID] = position
	}

	// Get all eternal farmings
	farmings, err := s.getAllEternalFarmings(farmingClient)
	if err != nil {
		return fmt.Errorf("failed to get eternal farmings: %w", err)
	}

	// Get all farming positions in one request
	allFarmingDeposits, err := s.getAllFarmingDeposits(farmingClient)
	if err != nil {
		return fmt.Errorf("failed to get farming positions: %w", err)
	}

	// Get all reward tokens info
	rewardTokenAddresses := make(map[string]bool)
	for _, farming := range farmings {
		rewardTokenAddresses[farming.RewardToken] = true
		if farming.BonusRewardToken != "0x0000000000000000000000000000000000000000" {
			rewardTokenAddresses[farming.BonusRewardToken] = true
		}
	}

	addresses := make([]string, 0, len(rewardTokenAddresses))
	for addr := range rewardTokenAddresses {
		addresses = append(addresses, addr)
	}

	var rewardTokens map[string]types.Token
	if len(addresses) > 0 {
		tokensList, err := s.getTokens(analyticsClient, addresses)
		if err != nil {
			return fmt.Errorf("failed to get tokens: %w", err)
		}
		rewardTokens = make(map[string]types.Token)
		for _, token := range tokensList {
			rewardTokens[token.ID] = token
		}
	}

	logger.Logger.Info("Fetched all data",
		zap.Int("pools", len(pools)),
		zap.Int("positions", len(positions)),
		zap.Int("farmings", len(farmings)),
		zap.Int("reward_tokens", len(rewardTokens)),
	)

	// Now process all calculations with the fetched data
	err = s.processPoolsAPR(pools, positions, poolFeesMap, networkID)
	if err != nil {
		logger.Logger.Error("Failed to process pools APR", zap.Error(err))
	}

	err = s.processPoolsMaxAPR(pools, positions, poolFeesMap, networkID)
	if err != nil {
		logger.Logger.Error("Failed to process pools max APR", zap.Error(err))
	}

	err = s.processFarmingsAPR(farmings, allFarmingDeposits, positionsById, rewardTokens, networkID)
	if err != nil {
		logger.Logger.Error("Failed to process farmings APR", zap.Error(err))
	}

	err = s.processFarmingsMaxAPR(farmings, allFarmingDeposits, positionsById, rewardTokens, networkID)
	if err != nil {
		logger.Logger.Error("Failed to process farmings max APR", zap.Error(err))
	}

	logger.Logger.Info("Completed full APR update", zap.String("network", network.Title))
	return nil
}

// Process pools APR calculation
func (s *APRService) processPoolsAPR(pools []types.Pool, positions []types.Position, poolFeesMap map[string]types.PoolDayData, networkID uint) error {
	logger.Logger.Info("Processing pools APR")

	// Group positions by pool ID for efficient lookup
	positionsByPool := make(map[string][]types.Position)
	for _, position := range positions {
		poolID := position.Pool.ID
		positionsByPool[poolID] = append(positionsByPool[poolID], position)
	}

	// Calculate APR for each pool
	for _, poolData := range pools {
		pool := s.findOrCreatePool(poolData, networkID)

		// Get positions for this pool
		poolPositions := positionsByPool[poolData.ID]

		// Calculate TVL and APR
		tvl := s.calculatePoolTVLFromPositions(poolData, poolPositions)
		fees := s.calculatePoolFeesFromData(poolData, poolFeesMap)

		if tvl > 0 {
			apr := (fees * 365 / tvl) * 100
			pool.LastAPR = &apr
		} else {
			apr := 0.0
			pool.LastAPR = &apr
		}

		s.db.Save(&pool)
	}

	logger.Logger.Info("Completed pools APR processing")
	return nil
}

// Process pools max APR calculation
func (s *APRService) processPoolsMaxAPR(pools []types.Pool, positions []types.Position, poolFeesMap map[string]types.PoolDayData, networkID uint) error {
	logger.Logger.Info("Processing pools max APR")

	// Group positions by pool ID
	positionsByPool := make(map[string][]types.Position)
	for _, position := range positions {
		poolID := position.Pool.ID
		positionsByPool[poolID] = append(positionsByPool[poolID], position)
	}

	// Calculate max APR for each pool
	for _, poolData := range pools {
		pool := s.findOrCreatePool(poolData, networkID)

		poolPositions := positionsByPool[poolData.ID]
		maxAPR := s.calculatePoolMaxAPRFromPositions(poolData, poolPositions, poolFeesMap)
		pool.MaxAPR = &maxAPR

		s.db.Save(&pool)
	}

	logger.Logger.Info("Completed pools max APR processing")
	return nil
}

// Process farmings APR calculation
func (s *APRService) processFarmingsAPR(farmings []types.EternalFarming, allFarmingDeposits []types.FarmingDeposit, positionsById map[string]types.Position, rewardTokens map[string]types.Token, networkID uint) error {
	logger.Logger.Info("Processing farmings APR")

	// Group farming positions by farming ID
	positionsByFarming := make(map[string][]types.Position)
	for _, farmingDeposit := range allFarmingDeposits {
		positionsByFarming[farmingDeposit.EternalFarming] = append(positionsByFarming[farmingDeposit.EternalFarming], positionsById[farmingDeposit.PositionID])
	}

	// Calculate APR for each farming
	for _, farmingData := range farmings {
		farming := s.findOrCreateEternalFarming(farmingData, networkID)

		farmingPositions := positionsByFarming[farmingData.ID]
		tvl := s.calculateFarmingActiveTVLFromPositions(farmingPositions)
		rewardRate := s.calculateFarmingRewardRateFromData(farmingData, rewardTokens)

		if tvl > 0 {
			apr := (rewardRate * 60 * 60 * 24 * 365 / tvl) * 100
			farming.LastAPR = &apr
		} else {
			apr := -1.0
			farming.LastAPR = &apr
		}

		farming.TVL = &tvl
		s.db.Save(&farming)
	}

	logger.Logger.Info("Completed farmings APR processing")
	return nil
}

// Process farmings max APR calculation
func (s *APRService) processFarmingsMaxAPR(farmings []types.EternalFarming, allFarmingDeposits []types.FarmingDeposit, positionsById map[string]types.Position, rewardTokens map[string]types.Token, networkID uint) error {
	logger.Logger.Info("Processing farmings max APR")

	// Group farming positions by farming ID
	positionsByFarming := make(map[string][]types.Position)
	for _, farmingDeposit := range allFarmingDeposits {
		positionsByFarming[farmingDeposit.EternalFarming] = append(positionsByFarming[farmingDeposit.EternalFarming], positionsById[farmingDeposit.PositionID])
	}

	// Calculate max APR for each farming
	for _, farmingData := range farmings {
		farming := s.findOrCreateEternalFarming(farmingData, networkID)

		farmingPositions := positionsByFarming[farmingData.ID]
		maxAPR := s.calculateFarmingMaxAPRFromPositions(farmingData, farmingPositions, rewardTokens)
		farming.MaxAPR = &maxAPR

		s.db.Save(&farming)
	}

	logger.Logger.Info("Completed farmings max APR processing")
	return nil
}

// Data fetching methods using GraphQL client
func (s *APRService) getAllPools(analyticsClient *client.GraphQLClient) ([]types.Pool, error) {
	var allPools []types.Pool
	const pageSize = 1000
	lastID := "0"

	for {
		variables := map[string]interface{}{
			"first": pageSize,
		}

		variables["id_gt"] = lastID

		result, err := analyticsClient.Execute(graphql.PoolsQuery, variables)
		if err != nil {
			return nil, err
		}

		var response types.PoolsResponse
		jsonData, _ := json.Marshal(result.Data)
		if err := json.Unmarshal(jsonData, &response); err != nil {
			return nil, err
		}

		if len(response.Pools) == 0 {
			break
		}

		allPools = append(allPools, response.Pools...)

		// Update lastID for next iteration
		lastID = response.Pools[len(response.Pools)-1].ID

		if len(response.Pools) < pageSize {
			break
		}
	}

	return allPools, nil
}

func (s *APRService) getPoolDayDatas(analyticsClient *client.GraphQLClient) ([]types.PoolDayData, error) {
	var allPoolDayDatas []types.PoolDayData
	const pageSize = 1000
	lastID := "0"

	// Get yesterday's timestamp in seconds
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayTimestamp := yesterday.Unix() / 86400 * 86400

	for {
		variables := map[string]interface{}{
			"date":  int(yesterdayTimestamp),
			"first": pageSize,
		}

		variables["id_gt"] = lastID

		result, err := analyticsClient.Execute(graphql.PoolDayDatasQuery, variables)
		if err != nil {
			return nil, err
		}

		var response types.PoolDayDatasResponse
		jsonData, _ := json.Marshal(result.Data)
		if err := json.Unmarshal(jsonData, &response); err != nil {
			return nil, err
		}

		if len(response.PoolDayDatas) == 0 {
			break
		}

		allPoolDayDatas = append(allPoolDayDatas, response.PoolDayDatas...)

		// Update lastID for next iteration
		lastID = response.PoolDayDatas[len(response.PoolDayDatas)-1].ID

		if len(response.PoolDayDatas) < pageSize {
			break
		}
	}

	return allPoolDayDatas, nil
}

func (s *APRService) getAllPositions(analyticsClient *client.GraphQLClient) ([]types.Position, error) {
	var allPositions []types.Position
	const pageSize = 1000
	lastID := "0"

	for {
		variables := map[string]interface{}{
			"first": pageSize,
		}

		variables["id_gt"] = lastID

		result, err := analyticsClient.Execute(graphql.PositionsQuery, variables)
		if err != nil {
			return nil, err
		}

		var response types.PositionsResponse
		jsonData, _ := json.Marshal(result.Data)
		if err := json.Unmarshal(jsonData, &response); err != nil {
			return nil, err
		}

		if len(response.Positions) == 0 {
			break
		}

		allPositions = append(allPositions, response.Positions...)

		// Update lastID for next iteration
		lastID = response.Positions[len(response.Positions)-1].ID

		if len(response.Positions) < pageSize {
			break
		}
	}

	return allPositions, nil
}

func (s *APRService) getAllEternalFarmings(farmingClient *client.GraphQLClient) ([]types.EternalFarming, error) {
	var allFarmings []types.EternalFarming
	const pageSize = 1000
	lastID := "0"

	for {
		variables := map[string]interface{}{
			"first": pageSize,
		}

		variables["id_gt"] = lastID

		result, err := farmingClient.Execute(graphql.FarmingsQuery, variables)
		if err != nil {
			return nil, err
		}

		var response types.EternalFarmingsResponse
		jsonData, _ := json.Marshal(result.Data)
		if err := json.Unmarshal(jsonData, &response); err != nil {
			return nil, err
		}

		if len(response.EternalFarmings) == 0 {
			break
		}

		allFarmings = append(allFarmings, response.EternalFarmings...)

		// Update lastID for next iteration
		lastID = response.EternalFarmings[len(response.EternalFarmings)-1].ID

		if len(response.EternalFarmings) < pageSize {
			break
		}
	}

	return allFarmings, nil
}

func (s *APRService) getAllFarmingDeposits(farmingClient *client.GraphQLClient) ([]types.FarmingDeposit, error) {
	var allFarmingDeposits []types.FarmingDeposit
	const pageSize = 1000
	lastID := "0"

	for {
		variables := map[string]interface{}{
			"first": pageSize,
		}

		variables["id_gt"] = lastID

		result, err := farmingClient.Execute(graphql.AllFarmingPositionsQuery, variables)
		if err != nil {
			return nil, err
		}

		var response types.AllFarmingPositionsResponse
		jsonData, _ := json.Marshal(result.Data)
		if err := json.Unmarshal(jsonData, &response); err != nil {
			return nil, err
		}

		if len(response.FarmingsDeposits) == 0 {
			break
		}

		allFarmingDeposits = append(allFarmingDeposits, response.FarmingsDeposits...)

		// Update lastID for next iteration
		lastID = response.FarmingsDeposits[len(response.FarmingsDeposits)-1].PositionID

		if len(response.FarmingsDeposits) < pageSize {
			break
		}
	}

	return allFarmingDeposits, nil
}

func (s *APRService) getTokens(analyticsClient *client.GraphQLClient, addresses []string) ([]types.Token, error) {
	variables := map[string]interface{}{
		"addresses": addresses,
	}

	result, err := analyticsClient.Execute(graphql.TokensQuery, variables)
	if err != nil {
		return nil, err
	}

	var response types.TokensResponse
	jsonData, _ := json.Marshal(result.Data)
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, err
	}

	return response.Tokens, nil
}

// Helper methods for finding/creating database records
func (s *APRService) findOrCreatePool(poolData types.Pool, networkID uint) models.Pool {
	address := poolData.ID

	var pool models.Pool
	result := s.db.Where("address = ? AND network_id = ?", address, networkID).First(&pool)

	if result.Error != nil {
		// Create new pool
		pool = models.Pool{
			Title:     fmt.Sprintf("%s : %s", poolData.Token0.Name, poolData.Token1.Name),
			Address:   address,
			NetworkID: networkID,
		}
		s.db.Create(&pool)
	}

	return pool
}

func (s *APRService) findOrCreateEternalFarming(farmingData types.EternalFarming, networkID uint) models.Farming {
	hash := farmingData.ID

	var farming models.Farming
	result := s.db.Where("hash = ? AND network_id = ?", hash, networkID).First(&farming)

	if result.Error != nil {
		// Create new farming
		farming = models.Farming{
			Hash:      hash,
			NetworkID: networkID,
		}
		s.db.Create(&farming)
	}

	return farming
}

// Calculation methods
func (s *APRService) calculatePoolTVLFromPositions(poolData types.Pool, positions []types.Position) float64 {
	totalTVL := 0.0

	tick, _ := strconv.Atoi(poolData.Tick)

	for _, position := range positions {
		liquidity, _ := strconv.ParseFloat(position.Liquidity, 64)
		token0Price, _ := strconv.ParseFloat(poolData.Token0Price, 64)
		tickLower, _ := strconv.Atoi(position.TickLower.TickIdx)
		tickUpper, _ := strconv.Atoi(position.TickUpper.TickIdx)

		// Check if position is in range
		if tickLower < tick && tick < tickUpper {
			amount0, amount1 := utils.GetAmounts(liquidity, tickLower, tickUpper, tick)

			// Convert to actual token amounts
			decimals0, _ := strconv.Atoi(poolData.Token0.Decimals)
			decimals1, _ := strconv.Atoi(poolData.Token1.Decimals)

			amount0 = amount0 / math.Pow(10, float64(decimals0))
			amount1 = amount1 / math.Pow(10, float64(decimals1))

			// Convert to native currency value
			totalTVL += amount0
			totalTVL += amount1 * token0Price
		}
	}

	return totalTVL
}

func (s *APRService) calculatePoolFeesFromData(poolData types.Pool, poolFeesMap map[string]types.PoolDayData) float64 {
	poolDayData, exists := poolFeesMap[poolData.ID]
	if !exists {
		return 0
	}

	feesToken0, _ := strconv.ParseFloat(poolDayData.FeesToken0, 64)
	feesToken1, _ := strconv.ParseFloat(poolDayData.FeesToken1, 64)
	token0Price, _ := strconv.ParseFloat(poolData.Token0Price, 64)

	return feesToken0 + feesToken1*token0Price
}

func (s *APRService) calculatePoolMaxAPRFromPositions(poolData types.Pool, positions []types.Position, poolFeesMap map[string]types.PoolDayData) float64 {
	maxAPR := 0.0

	tick, _ := strconv.Atoi(poolData.Tick)
	totalLiquidity, _ := strconv.ParseFloat(poolData.Liquidity, 64)
	token0Price, _ := strconv.ParseFloat(poolData.Token0Price, 64)
	totalFees := s.calculatePoolFeesFromData(poolData, poolFeesMap)

	for _, position := range positions {
		liquidity, _ := strconv.ParseFloat(position.Liquidity, 64)
		tickLower, _ := strconv.Atoi(position.TickLower.TickIdx)
		tickUpper, _ := strconv.Atoi(position.TickUpper.TickIdx)

		// Check if position is in range
		if tickLower < tick && tick < tickUpper {
			amount0, amount1 := utils.GetAmounts(liquidity, tickLower, tickUpper, tick)

			// Convert to actual token amounts
			decimals0, _ := strconv.Atoi(poolData.Token0.Decimals)
			decimals1, _ := strconv.Atoi(poolData.Token1.Decimals)

			amount0 = amount0 / math.Pow(10, float64(decimals0))
			amount1 = amount1 / math.Pow(10, float64(decimals1))

			positionTVL := amount0 + amount1*token0Price
			positionFees := totalFees * liquidity / totalLiquidity

			if positionTVL > 0 {
				apr := (positionFees * 365 / positionTVL) * 100
				if apr > maxAPR {
					maxAPR = apr
				}
			}
		}
	}

	return maxAPR
}

func (s *APRService) calculateFarmingActiveTVLFromPositions(positions []types.Position) float64 {
	activeTVL := 0.0

	for _, position := range positions {
		tick, _ := strconv.Atoi(position.Pool.Tick)
		liquidity, _ := strconv.ParseFloat(position.Liquidity, 64)
		tickLower, _ := strconv.Atoi(position.TickLower.TickIdx)
		tickUpper, _ := strconv.Atoi(position.TickUpper.TickIdx)

		if tick > tickLower && tick < tickUpper {

			amount0, amount1 := utils.GetAmounts(liquidity, tickLower, tickUpper, tick)

			// Convert to actual token amounts
			decimals0, _ := strconv.Atoi(position.Pool.Token0.Decimals)
			decimals1, _ := strconv.Atoi(position.Pool.Token1.Decimals)
			derivedMatic0, _ := strconv.ParseFloat(position.Pool.Token0.DerivedMatic, 64)
			derivedMatic1, _ := strconv.ParseFloat(position.Pool.Token1.DerivedMatic, 64)

			amount0 = amount0 / math.Pow(10, float64(decimals0))
			amount1 = amount1 / math.Pow(10, float64(decimals1))

			// Convert to native currency value
			activeTVL += amount0 * derivedMatic0
			activeTVL += amount1 * derivedMatic1
		}
	}

	return activeTVL
}

func (s *APRService) calculateFarmingRewardRateFromData(farmingData types.EternalFarming, tokens map[string]types.Token) float64 {
	rewardRate := 0.0

	// Main reward token
	if token, exists := tokens[farmingData.RewardToken]; exists {
		rate, _ := strconv.ParseFloat(farmingData.RewardRate, 64)
		decimals, _ := strconv.Atoi(token.Decimals)
		derivedMatic, _ := strconv.ParseFloat(token.DerivedMatic, 64)

		rewardRate += (rate / math.Pow(10, float64(decimals))) * derivedMatic
	}

	// Bonus reward token
	if farmingData.BonusRewardToken != "0x0000000000000000000000000000000000000000" {
		if token, exists := tokens[farmingData.BonusRewardToken]; exists {
			rate, _ := strconv.ParseFloat(farmingData.BonusRewardRate, 64)
			decimals, _ := strconv.Atoi(token.Decimals)
			derivedMatic, _ := strconv.ParseFloat(token.DerivedMatic, 64)

			rewardRate += (rate / math.Pow(10, float64(decimals))) * derivedMatic
		}
	}

	return rewardRate
}

func (s *APRService) calculateFarmingMaxAPRFromPositions(farmingData types.EternalFarming, positions []types.Position, tokens map[string]types.Token) float64 {
	maxAPR := 0.0

	rewardRate := s.calculateFarmingRewardRateFromData(farmingData, tokens)
	totalActiveLiquidity := 0.0

	// Calculate total active liquidity
	for _, position := range positions {
		tick, _ := strconv.Atoi(position.Pool.Tick)
		tickLower, _ := strconv.Atoi(position.TickLower.TickIdx)
		tickUpper, _ := strconv.Atoi(position.TickUpper.TickIdx)

		if tickLower < tick && tick < tickUpper {
			liquidity, _ := strconv.ParseFloat(position.Liquidity, 64)
			totalActiveLiquidity += liquidity
		}
	}

	// Calculate max APR for each position
	for _, position := range positions {
		tick, _ := strconv.Atoi(position.Pool.Tick)
		liquidity, _ := strconv.ParseFloat(position.Liquidity, 64)
		tickLower, _ := strconv.Atoi(position.TickLower.TickIdx)
		tickUpper, _ := strconv.Atoi(position.TickUpper.TickIdx)

		if tickLower < tick && tick < tickUpper {
			amount0, amount1 := utils.GetAmounts(liquidity, tickLower, tickUpper, tick)

			// Convert to actual token amounts
			decimals0, _ := strconv.Atoi(position.Pool.Token0.Decimals)
			decimals1, _ := strconv.Atoi(position.Pool.Token1.Decimals)
			derivedMatic0, _ := strconv.ParseFloat(position.Pool.Token0.DerivedMatic, 64)
			derivedMatic1, _ := strconv.ParseFloat(position.Pool.Token1.DerivedMatic, 64)

			amount0 = amount0 / math.Pow(10, float64(decimals0))
			amount1 = amount1 / math.Pow(10, float64(decimals1))

			positionTVL := amount0*derivedMatic0 + amount1*derivedMatic1

			if positionTVL > 0 && totalActiveLiquidity > 0 {
				positionRewardRate := rewardRate * liquidity / totalActiveLiquidity
				apr := (positionRewardRate * 60 * 60 * 24 * 365 / positionTVL) * 100

				if apr > maxAPR {
					maxAPR = apr
				}
			}
		}
	}

	return maxAPR
}
