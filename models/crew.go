package models

import "time"

type Crew struct {
	CrewID             string    `json:"crewID" dynamodbav:"crewID"`
	CreatedAt          time.Time `json:"createdAt" dynamodbav:"createdAt"`
	CreatedBy          string    `json:"createdBy" dynamodbav:"createdBy" validate:"required"`
	Description        string    `json:"description" dynamodbav:"description" validate:"omitempty,max=500"`
	IsActive           bool      `json:"isActive" dynamodbav:"isActive"`
	LeadTechnicianId   string    `json:"leadTechnicianId" dynamodbav:"leadTechnicianId" validate:"required"`
	MemberIds          []string  `json:"memberIds" dynamodbav:"memberIds"`
	Name               string    `json:"name" dynamodbav:"name" validate:"required,min=2,max=100"`
	OrgID              string    `json:"orgID" dynamodbav:"orgID" validate:"required"`
	Skills             []string  `json:"skills" dynamodbav:"skills"`
	UpdatedAt          time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}

type CreateCrewRequest struct {
	Description      string   `json:"description" validate:"omitempty,max=500"`
	LeadTechnicianId string   `json:"leadTechnicianId" validate:"required"`
	MemberIds        []string `json:"memberIds"`
	Name             string   `json:"name" validate:"required,min=2,max=100"`
	OrgID            string   `json:"orgID" validate:"required"`
	Skills           []string `json:"skills"`
}

type UpdateCrewRequest struct {
	Description      string   `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive         *bool    `json:"isActive,omitempty"`
	LeadTechnicianId string   `json:"leadTechnicianId,omitempty"`
	MemberIds        []string `json:"memberIds,omitempty"`
	Name             string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Skills           []string `json:"skills,omitempty"`
}

type CrewFilter struct {
	OrgID            string `json:"orgID,omitempty"`
	LeadTechnicianId string `json:"leadTechnicianId,omitempty"`
	IsActive         *bool  `json:"isActive,omitempty"`
	CreatedBy        string `json:"createdBy,omitempty"`
	FromDate         time.Time `json:"fromDate,omitempty"`
	ToDate           time.Time `json:"toDate,omitempty"`
}