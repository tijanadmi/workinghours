package models

import (
	"time"
)

// User is the user model
type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Password  string
	UserRole  []UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRole struct {
	ID        int
	UserID    int
	OrgID     int
	OrgUnit   OrgUnit
	CreatedAt time.Time
	UpdatedAt time.Time
	RoleType  string
}

type OrgUnit struct {
	ID        int
	Code      string
	Name      string
	OrgID     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Shift struct {
	ID        int
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Employee struct {
	ID        int
	Code      string
	FirstName string
	LastName  string
	Workplace string
	Org1ID    int
	Org2ID    int
	OrgUnit1  string
	OrgUnit2  string
	Location  string
	Address   string
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EmpDaysReservation struct {
	ID           int
	StartDate    time.Time
	ShiftID      int
	EmpID        int
	UserCreateID int
	Shift        Shift
	Employee     Employee
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ROOM is the room model
type Room struct {
	ID        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Restriction is the restriction model
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservation is the reservation model
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room
	Processed int
}

// RoomRestriction is the room restriction model
type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

//MailData holds an email message
type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}
