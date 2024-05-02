package types

type PathClassification int

const (
	PathClassificationNone PathClassification = iota
	PathClassificationAncestor
	PathClassificationCourse
	PathClassificationDescendant
)
