package schema

type GoalType string

const (
	GoalTypePersonal GoalType = "personal"
	GoalTypeGlobal   GoalType = "global"
	GoalTypeShared   GoalType = "shared"
)

type GoalKind string

const (
	GoalKindQualitative  GoalKind = "qualitative"
	GoalKindQuantitative GoalKind = "quantitative"
)
