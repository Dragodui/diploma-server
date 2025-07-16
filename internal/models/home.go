package models

import "time"

type Home struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:64;not null" json:"name"`
	InviteCode string    `gorm:"size:64;not null;unique" json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`

	// relations
	Memberships     []HomeMembership `gorm:"foreignKey:HomeID" json:"memberships,omitempty"`
	Tasks           []Task           `gorm:"foreignKey:HomeID" json:"tasks,omitempty"`
	TaskAssignments []TaskAssignment `gorm:"foreignKey:HomeID" json:"task_assignments,omitempty"`
}

type CreateHomeRequest struct {
	Name string `json:"name" validate:"required,min=8"`
}
