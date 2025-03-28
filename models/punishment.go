package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

// Punishment
// a record of user punishment
// when a record created, it can't be modified if other admins punish this user on the same floor
// whether a user is banned to post on one division based on the latest / max(id) record
// if admin want to modify punishment duration, manually modify the latest record of this user in database
// admin can be granted update privilege on SQL view of this table
type Punishment struct {
	ID int `json:"id" gorm:"primaryKey"`

	// time when this punishment creates
	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`

	// time when this punishment revoked
	DeletedAt gorm.DeletedAt `json:"-"`

	// start from end_time of previous punishment (punishment accumulation of different floors)
	// if no previous punishment or previous punishment end time less than time.Now() (synced), set start time time.Now()
	StartTime time.Time `json:"start_time" gorm:"not null"`

	// end_time of this punishment
	EndTime time.Time `json:"end_time" gorm:"not null"`

	Duration *time.Duration `json:"duration" swaggertype:"integer"`

	Day int `json:"day"`

	// user punished
	UserID int `json:"user_id" gorm:"not null;index"`

	// admin user_id who made this punish
	MadeBy int `json:"made_by,omitempty"`

	// punished because of this floor
	FloorID *int `json:"floor_id" gorm:"uniqueIndex"`

	Floor *Floor `json:"floor,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // foreign key

	DivisionID int `json:"division_id" gorm:"not null"`

	Division *Division `json:"division,omitempty"` // foreign key

	// reason
	Reason string `json:"reason" gorm:"size:128"`
}

type Punishments []*Punishment

func (punishment *Punishment) Create() (*User, error) {
	var user User

	err := DB.Clauses(dbresolver.Write).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&user, punishment.UserID).Error
		if err != nil {
			return err
		}

		var previousPunishment Punishment
		err = tx.Where("user_id = ? and floor_id = ?", user.ID, punishment.FloorID).Take(&previousPunishment).Error
		if err == nil {
			// return common.Forbidden("该用户已被禁言")

			// same as before, do nothing
			if previousPunishment.Duration == punishment.Duration && previousPunishment.Day == punishment.Day {
				return nil
			}

			// different duration, revoke previous punishment
			diffDuration := time.Duration(punishment.Day-previousPunishment.Day) * 24 * time.Hour

			previousPunishment.Duration = punishment.Duration
			previousPunishment.Day = punishment.Day
			previousPunishment.EndTime = previousPunishment.StartTime.Add(*punishment.Duration)
			previousPunishment.Reason = punishment.Reason
			previousPunishment.MadeBy = punishment.MadeBy
			// conflict with previous punishment if not equal
			// ignore it as it's rare
			previousPunishment.DivisionID = punishment.DivisionID

			if user.BanDivision[punishment.DivisionID] == nil {
				user.BanDivision[punishment.DivisionID] = &previousPunishment.EndTime
			} else {
				*user.BanDivision[punishment.DivisionID] = user.BanDivision[punishment.DivisionID].Add(diffDuration)
			}

			err = tx.Updates(&previousPunishment).Error
			if err != nil {
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		} else {
			user.OffenceCount += 1
			punishment.StartTime = time.Now()
			punishment.EndTime = punishment.StartTime.Add(*punishment.Duration)

			if user.BanDivision[punishment.DivisionID] == nil {
				user.BanDivision[punishment.DivisionID] = &punishment.EndTime
			} else {
				*user.BanDivision[punishment.DivisionID] = user.BanDivision[punishment.DivisionID].Add(*punishment.Duration)
			}

			err = tx.Create(&punishment).Error
			if err != nil {
				return err
			}
		}

		err = tx.Select("BanDivision", "OffenceCount").Save(&user).Error
		if err != nil {
			return err
		}

		return nil
	})
	return &user, err
}
