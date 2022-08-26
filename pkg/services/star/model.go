package star

import "errors"

var ErrCommandValidationFailed = errors.New("command missing required fields")

type Star struct {
	ID          int64 `xorm:"pk autoincr 'id'" db:"id"`
	UserID      int64 `xorm:"user_id" db:"user_id"`
	DashboardID int64 `xorm:"dashboard_id" db:"dashboard_id"`
}

// ----------------------
// COMMANDS

type StarDashboardCommand struct {
	UserID       int64  `xorm:"user_id"`
	DashboardID  int64  `xorm:"dashboard_id"`
	DashboardUID string `xorm:"-"`
}

func (cmd *StarDashboardCommand) Validate() error {
	if cmd.DashboardID == 0 && cmd.DashboardUID == "" || cmd.UserID == 0 {
		return ErrCommandValidationFailed
	}
	return nil
}

type UnstarDashboardCommand struct {
	UserID       int64  `xorm:"user_id"`
	DashboardID  int64  `xorm:"dashboard_id"`
	DashboardUID string `xorm:"-"`
}

func (cmd *UnstarDashboardCommand) Validate() error {
	if cmd.DashboardID == 0 && cmd.DashboardUID == "" || cmd.UserID == 0 {
		return ErrCommandValidationFailed
	}
	return nil
}

// ---------------------
// QUERIES

type GetUserStarsQuery struct {
	OrgID  int64 `xorm:"-"`
	UserID int64 `xorm:"user_id"`
}

type IsStarredByUserQuery struct {
	UserID       int64  `xorm:"user_id"`
	DashboardID  int64  `xorm:"dashboard_id"`
	DashboardUID string `xorm:"-"`
}

type GetUserStarsResult struct {
	UserStars    map[int64]bool
	UserStarsUID []string `json:"userStarsUid"`
}
