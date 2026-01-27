package types

// GraphQL response types
type Token struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     string `json:"decimals"`
	DerivedMatic string `json:"derivedMatic"`
}

type Pool struct {
	ID          string `json:"id"`
	Tick        string `json:"tick"`
	Token0      Token  `json:"token0"`
	Token1      Token  `json:"token1"`
	Token0Price string `json:"token0Price"`
	Liquidity   string `json:"liquidity"`
	FeesToken0  string `json:"feesToken0"`
	FeesToken1  string `json:"feesToken1"`
}

type Tick struct {
	TickIdx string `json:"tickIdx"`
}

type Position struct {
	ID        string `json:"id"`
	Liquidity string `json:"liquidity"`
	TickLower Tick   `json:"tickLower"`
	TickUpper Tick   `json:"tickUpper"`
	Pool      Pool   `json:"pool"`
	Owner     string `json:"owner"`
}

type EternalFarming struct {
	ID               string `json:"id"`
	RewardToken      string `json:"rewardToken"`
	BonusRewardToken string `json:"bonusRewardToken"`
	RewardRate       string `json:"rewardRate"`
	BonusRewardRate  string `json:"bonusRewardRate"`
	Pool             string `json:"pool"`
}

type FarmingDeposit struct {
	PositionID     string `json:"id"`
	EternalFarming string `json:"eternalFarming"`
}

type PoolDayData struct {
	ID         string `json:"id"`
	FeesToken0 string `json:"feesToken0"`
	FeesToken1 string `json:"feesToken1"`
	Date       int64  `json:"date"`
	Pool       struct {
		ID string `json:"id"`
	} `json:"pool"`
}

// Response structures
type PoolsResponse struct {
	Pools []Pool `json:"pools"`
}

type PositionsResponse struct {
	Positions []Position `json:"positions"`
}

type EternalFarmingsResponse struct {
	EternalFarmings []EternalFarming `json:"eternalFarmings"`
}

type AllFarmingPositionsResponse struct {
	FarmingsDeposits []FarmingDeposit `json:"deposits"`
}

type TokensResponse struct {
	Tokens []Token `json:"tokens"`
}

type PoolDayDatasResponse struct {
	PoolDayDatas []PoolDayData `json:"poolDayDatas"`
}
