package models

import "time"

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusActive     JobStatus = "active"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusCancelled  JobStatus = "cancelled"
	JobStatusOnHold     JobStatus = "on_hold"
)

type JobType string

const (
	JobTypeService      JobType = "service"
	JobTypeMaintenance  JobType = "maintenance"
	JobTypeInstallation JobType = "installation"
	JobTypeRepair       JobType = "repair"
	JobTypeInspection   JobType = "inspection"
)

type CreatedData struct {
	UID        string `json:"uID" dynamodbav:"uID"`
	UserName   string `json:"userName" dynamodbav:"userName"`
	UserStatus string `json:"userStatus" dynamodbav:"userStatus"`
}

type StartedData struct {
	UID        string    `json:"uID,omitempty" dynamodbav:"uID,omitempty"`
	UserName   string    `json:"userName,omitempty" dynamodbav:"userName,omitempty"`
	UserStatus string    `json:"userStatus,omitempty" dynamodbav:"userStatus,omitempty"`
	StartedAt  time.Time `json:"startedAt,omitempty" dynamodbav:"startedAt,omitempty"`
}

type DeletedData struct {
	UID        string    `json:"uID,omitempty" dynamodbav:"uID,omitempty"`
	UserName   string    `json:"userName,omitempty" dynamodbav:"userName,omitempty"`
	UserStatus string    `json:"userStatus,omitempty" dynamodbav:"userStatus,omitempty"`
	DeletedAt  time.Time `json:"deletedAt,omitempty" dynamodbav:"deletedAt,omitempty"`
	Reason     string    `json:"reason,omitempty" dynamodbav:"reason,omitempty"`
}

type QBInfoOnJob struct {
	CustomerID  string `json:"customerID,omitempty" dynamodbav:"customerID,omitempty"`
	InvoiceID   string `json:"invoiceID,omitempty" dynamodbav:"invoiceID,omitempty"`
	LineItemID  string `json:"lineItemID,omitempty" dynamodbav:"lineItemID,omitempty"`
	PaymentID   string `json:"paymentID,omitempty" dynamodbav:"paymentID,omitempty"`
	ScheduledAt string `json:"scheduledAt,omitempty" dynamodbav:"scheduledAt,omitempty"`
	ServiceNotes string `json:"serviceNotes,omitempty" dynamodbav:"serviceNotes,omitempty"`
}

type Job struct {
	JobID                   string      `json:"jobID" dynamodbav:"jobID" validate:"omitempty,uuid4"`
	ClientID                string      `json:"clientID" dynamodbav:"clientID" validate:"required"`
	CreatedAt               time.Time   `json:"createdAt" dynamodbav:"createdAt"`
	CreatedData             CreatedData `json:"createdData" dynamodbav:"createdData" validate:"required"`
	DeletedData             *DeletedData `json:"deletedData,omitempty" dynamodbav:"deletedData,omitempty"`
	InvID                   string      `json:"invID,omitempty" dynamodbav:"invID,omitempty"`
	JobEndedAt              *time.Time  `json:"jobEndedAt,omitempty" dynamodbav:"jobEndedAt,omitempty"`
	JobImagesAfterService   []string    `json:"jobImagesAfterService" dynamodbav:"jobImagesAfterService"`
	JobsName                string      `json:"jobsName" dynamodbav:"jobsName" validate:"required,min=2,max=200"`
	JobStartedAt            *time.Time  `json:"jobStartedAt,omitempty" dynamodbav:"jobStartedAt,omitempty"`
	JobStatus               JobStatus   `json:"jobStatus" dynamodbav:"jobStatus" validate:"required,oneof=pending active in_progress completed cancelled on_hold"`
	JobType                 JobType     `json:"jobType" dynamodbav:"jobType" validate:"required,oneof=service maintenance installation repair inspection"`
	Notes                   string      `json:"notes,omitempty" dynamodbav:"notes,omitempty" validate:"omitempty,max=1000"`
	OrgID                   string      `json:"orgID" dynamodbav:"orgID" validate:"required"`
	PaymentID               string      `json:"paymentID,omitempty" dynamodbav:"paymentID,omitempty"`
	QBInfoOnJob             *QBInfoOnJob `json:"qbInfoOnJob,omitempty" dynamodbav:"qbInfoOnJob,omitempty"`
	StartedData             *StartedData `json:"startedData,omitempty" dynamodbav:"startedData,omitempty"`
	UsersAssignedToJob      []string    `json:"usersAssignedToJob" dynamodbav:"usersAssignedToJob"`
	VehiclesAssignedToJob   []string    `json:"vehiclesAssignedToJob" dynamodbav:"vehiclesAssignedToJob"`
	UpdatedAt               time.Time   `json:"updatedAt,omitempty" dynamodbav:"updatedAt,omitempty"`
	UpdatedBy               string      `json:"updatedBy,omitempty" dynamodbav:"updatedBy,omitempty"`
}

type CreateJobRequest struct {
	ClientID              string      `json:"clientID" validate:"required"`
	JobsName              string      `json:"jobsName" validate:"required,min=2,max=200"`
	JobType               JobType     `json:"jobType" validate:"required,oneof=service maintenance installation repair inspection"`
	Notes                 string      `json:"notes,omitempty" validate:"omitempty,max=1000"`
	OrgID                 string      `json:"orgID" validate:"required"`
	UsersAssignedToJob    []string    `json:"usersAssignedToJob,omitempty"`
	VehiclesAssignedToJob []string    `json:"vehiclesAssignedToJob,omitempty"`
	QBInfoOnJob           *QBInfoOnJob `json:"qbInfoOnJob,omitempty"`
}

type UpdateJobRequest struct {
	JobsName              string      `json:"jobsName,omitempty" validate:"omitempty,min=2,max=200"`
	JobStatus             JobStatus   `json:"jobStatus,omitempty" validate:"omitempty,oneof=pending active in_progress completed cancelled on_hold"`
	JobType               JobType     `json:"jobType,omitempty" validate:"omitempty,oneof=service maintenance installation repair inspection"`
	Notes                 string      `json:"notes,omitempty" validate:"omitempty,max=1000"`
	UsersAssignedToJob    []string    `json:"usersAssignedToJob,omitempty"`
	VehiclesAssignedToJob []string    `json:"vehiclesAssignedToJob,omitempty"`
	QBInfoOnJob           *QBInfoOnJob `json:"qbInfoOnJob,omitempty"`
	JobImagesAfterService []string    `json:"jobImagesAfterService,omitempty"`
}

type JobFilter struct {
	OrgID     string    `json:"orgID,omitempty"`
	ClientID  string    `json:"clientID,omitempty"`
	JobStatus JobStatus `json:"jobStatus,omitempty"`
	JobType   JobType   `json:"jobType,omitempty"`
	CreatedBy string    `json:"createdBy,omitempty"`
	FromDate  time.Time `json:"fromDate,omitempty"`
	ToDate    time.Time `json:"toDate,omitempty"`
}