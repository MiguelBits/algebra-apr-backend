package graphql

import (
	_ "embed"
)

//go:embed pools.graphql
var PoolsQuery string

//go:embed positions.graphql
var PositionsQuery string

//go:embed farmings.graphql
var FarmingsQuery string

//go:embed all_farming_positions.graphql
var AllFarmingPositionsQuery string

//go:embed tokens.graphql
var TokensQuery string

//go:embed pool_day_datas.graphql
var PoolDayDatasQuery string
