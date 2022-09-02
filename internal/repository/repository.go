package repository

import (
	"time"

	"github.com/tijanadmi/workinghours/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool
	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	AllRooms() ([]models.Room, error)
	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlockByID(id int) error
	InsertReservationDayForEmp(shift_id int, emp_id int, user_create_id int, startDate time.Time) error
	DeleteEmpDayByID(id int) error
}
