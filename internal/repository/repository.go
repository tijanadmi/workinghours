package repository

import (
	"time"

	"github.com/tijanadmi/workinghours/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool
	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	//Authenticate(email, testPassword string) (int, string, error)
	Authenticate(email, testPassword string) (int, string, []int, []int, error)

	AllRooms() ([]models.Room, error)
	GetEmployeeByUserIDCRUD(user_id int) ([]models.Employee, error)
	GetEmployeeByUserIDGLE(user_id int) ([]models.Employee, error)
	GetEmployeeByOrgID(org_id int) ([]models.Employee, error)
	GetReservationEmployeeByDate(shiftID, org_id int, start time.Time) ([]models.Employee, error)
	GetWorkingDayTypes() ([]models.WorkingDayType, error)
	GetOrgUnitsByUserIDGLE(user_id int) ([]models.OrgUnit, error)
	GetReservationForEmpByDate(shiftID int, empID int, start, end time.Time) ([]models.EmpDaysReservation, error)
	GetReservationTypeForEmpByDate(empID int, start, end time.Time) ([]models.EmpDaysReservation, error)
	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlockByID(id int) error
	InsertReservationDayForEmp(shift_id int, emp_id int, user_create_id int, startDate time.Time) error
	InsertReservationDayTypeForEmp(wd_type_id int, emp_id int, user_create_id int, startDate time.Time) error
	DeleteEmpDayByID(id int) error
	DeleteEmpDayTypeByID(wd_type_id int, emp_id int, user_create_id int, startDate time.Time) error
}
