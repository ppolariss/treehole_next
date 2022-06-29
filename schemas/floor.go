package schemas

type HoleID struct {
	HoleID int `json:"hole_id"`
}

type ListFloorOld struct {
	HoleID
	Size   int `json:"length" default:"20"`
	Offset int `json:"start_floor"  default:"0"`
}

type CreateFloor struct {
	// Owner or admin, if it's modified or deleted, the original content should be moved to  floor_history
	Content string `json:"content"`
	// Admin only
	SpecialTag string `json:"special_tag" default:""`
}

type CreateFloorOld struct {
	HoleID
	CreateFloor
}

type ModifyFloor struct {
	CreateFloor
	// All users, 1 is like, -1 is dislike, 0 is reset
	Like int `json:"like_int"`
	// To be compatible with the deprecated API, "add" is like, "cancel" is reset
	LikeOld string `json:"like"`
	// Admin only
	Fold string `json:"fold"`
}

type DeleteFloor struct {
	Reason string `json:"delete_reason"`
}