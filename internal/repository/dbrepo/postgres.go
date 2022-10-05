package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/tijanadmi/workinghours/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// InsertReservation inserts a reservation into the database

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// GetUserByID returns a user by id
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, password,  created_at, updated_at
			from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

// UpdateUser updates a user in the database
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email = $3,  updated_at = $5
`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, []int, []int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", nil, nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", nil, nil, errors.New("incorrect password")
	} else if err != nil {
		return 0, "", nil, nil, err
	}

	/********/
	query := `select ur.id, ur.user_id,  o.id, o.code, o.name, ur.role_type
	from user_roles ur
	left join org_units o on o.id=ur.org_id 
	where ur.user_id  = $1
	`

	var crudRole []int
	var gleRole []int
	rows, _ := m.DB.QueryContext(ctx, query, id)
	defer rows.Close()

	for rows.Next() {
		var ur models.UserRole
		err := rows.Scan(
			&ur.ID,
			&ur.UserID,
			&ur.OrgID,
			&ur.OrgUnit.Code,
			&ur.OrgUnit.Name,
			&ur.RoleType,
		)

		if err != nil {
			return 0, "", nil, nil, err
		}
		if ur.RoleType == "CRUD" {
			crudRole = append(crudRole, ur.OrgID)
		} else {
			gleRole = append(gleRole, ur.OrgID)
		}

	}
	/********/

	return id, hashedPassword, crudRole, gleRole, nil
}

func (m *postgresDBRepo) GetEmployeeByUserIDCRUD(user_id int) ([]models.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var employee []models.Employee

	query := `select id, code, first_name, last_name, workplace, org1_id, coalesce(org2_id, 0), org_unit1,  location, address, phone, email, created_at, updated_at 
			  from employee where org1_id in (select org_id from user_roles where user_id=$1 and role_type='CRUD')  order by org1_id, first_name`

	rows, err := m.DB.QueryContext(ctx, query, user_id)

	if err != nil {
		return employee, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Employee
		err := rows.Scan(
			&e.ID,
			&e.Code,
			&e.FirstName,
			&e.LastName,
			&e.Workplace,
			&e.Org1ID,
			&e.Org2ID,
			&e.OrgUnit1,
			&e.Location,
			&e.Address,
			&e.Phone,
			&e.Email,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return employee, err
		}
		employee = append(employee, e)
	}

	if err = rows.Err(); err != nil {
		return employee, err
	}

	return employee, nil

}

func (m *postgresDBRepo) GetEmployeeByUserIDGLE(user_id int) ([]models.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var employee []models.Employee

	query := `select id, code, first_name, last_name, workplace, org1_id, coalesce(org2_id, 0), org_unit1,  location, address, phone, email, created_at, updated_at 
			  from employee where org1_id in (select org_id from user_roles where user_id=$1 and role_type='GLE')  order by org1_id, first_name`

	rows, err := m.DB.QueryContext(ctx, query, user_id)

	if err != nil {
		return employee, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Employee
		err := rows.Scan(
			&e.ID,
			&e.Code,
			&e.FirstName,
			&e.LastName,
			&e.Workplace,
			&e.Org1ID,
			&e.Org2ID,
			&e.OrgUnit1,
			&e.Location,
			&e.Address,
			&e.Phone,
			&e.Email,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return employee, err
		}
		employee = append(employee, e)
	}

	if err = rows.Err(); err != nil {
		return employee, err
	}

	return employee, nil

}

func (m *postgresDBRepo) GetEmployeeByOrgID(org_id int) ([]models.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var employee []models.Employee

	query := `select id, code, first_name, last_name, workplace, org1_id, coalesce(org2_id, 0), org_unit1,  location, address, phone, email, created_at, updated_at 
			  from employee where org1_id =$1  order by org1_id, first_name`

	rows, err := m.DB.QueryContext(ctx, query, org_id)

	if err != nil {
		return employee, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Employee
		err := rows.Scan(
			&e.ID,
			&e.Code,
			&e.FirstName,
			&e.LastName,
			&e.Workplace,
			&e.Org1ID,
			&e.Org2ID,
			&e.OrgUnit1,
			&e.Location,
			&e.Address,
			&e.Phone,
			&e.Email,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return employee, err
		}
		employee = append(employee, e)
	}

	if err = rows.Err(); err != nil {
		return employee, err
	}

	return employee, nil

}

func (m *postgresDBRepo) GetOrgUnitsByUserIDGLE(user_id int) ([]models.OrgUnit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orgUnits []models.OrgUnit

	query := `select id, code, name, created_at, updated_at 
			  from org_units where id in (select org_id from user_roles where user_id=$1 and role_type='GLE')  order by code`

	rows, err := m.DB.QueryContext(ctx, query, user_id)

	if err != nil {
		return orgUnits, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.OrgUnit
		err := rows.Scan(
			&e.ID,
			&e.Code,
			&e.Name,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return orgUnits, err
		}
		orgUnits = append(orgUnits, e)
	}

	if err = rows.Err(); err != nil {
		return orgUnits, err
	}

	return orgUnits, nil

}

func (m *postgresDBRepo) GetWorkingDayTypes() ([]models.WorkingDayType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var workingDayTypes []models.WorkingDayType

	query := `select id, code, name, created_at, updated_at 
			  from working_day_types order by code`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return workingDayTypes, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.WorkingDayType
		err := rows.Scan(
			&e.ID,
			&e.Code,
			&e.Name,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return workingDayTypes, err
		}
		workingDayTypes = append(workingDayTypes, e)
	}

	if err = rows.Err(); err != nil {
		return workingDayTypes, err
	}

	return workingDayTypes, nil

}

func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `
		select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
		from room_restrictions where $1 < end_date and $2 >= start_date
		and room_id = $3
`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return nil, err
		}
		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}

// InsertBlockForRoom inserts a room restriction
func (m *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into room_restrictions (start_date, end_date, room_id, restriction_id,
			created_at, updated_at) values ($1, $2, $3, $4, $5, $6)`

	_, err := m.DB.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// DeleteBlockByID deletes a room restriction
func (m *postgresDBRepo) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from room_restrictions where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetReservationForEmpByDate returns reservations for a emp and shift by date range
func (m *postgresDBRepo) GetReservationEmployeeByDate(shiftID, org_id int, start time.Time) ([]models.Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var employee []models.Employee
	query := `
		select coalesce(id, 0), first_name
		from employee 
		where org1_id = $1
		and id in (select emp_id from emp_days_reservations where  start_date = $2 and shift_id = $3)
`

	rows, err := m.DB.QueryContext(ctx, query, org_id, start, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Employee
		err := rows.Scan(
			&e.ID,
			&e.FirstName,
		)
		if err != nil {
			return nil, err
		}
		employee = append(employee, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return employee, nil
}

// GetReservationForEmpByDate returns reservations for a emp and shift by date range
func (m *postgresDBRepo) GetReservationForEmpByDate(shiftID int, empID int, start, end time.Time) ([]models.EmpDaysReservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.EmpDaysReservation
	query := `
		select id, start_date, shift_id,emp_id,user_create_id
		from emp_days_reservations where $3 <= start_date and $4 >= start_date
		and shift_id = $1
		and emp_id = $2
`

	rows, err := m.DB.QueryContext(ctx, query, shiftID, empID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.EmpDaysReservation
		err := rows.Scan(
			&r.ID,
			&r.StartDate,
			&r.ShiftID,
			&r.EmpID,
			&r.UserCreateID,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

// InsertReservationDayForEmp inserts reservation day for emp
func (m *postgresDBRepo) InsertReservationDayForEmp(shift_id int, emp_id int, user_create_id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into emp_days_reservations (start_date, shift_id, emp_id, user_create_id,
			created_at, updated_at) values ($1, $2, $3, $4, $5, $6)`
	log.Println(shift_id)
	_, err := m.DB.ExecContext(ctx, query, startDate, shift_id, emp_id, user_create_id, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// DeleteEmpDayByID deletes a emp day reservation
func (m *postgresDBRepo) DeleteEmpDayByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from emp_days_reservations where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetReservationForEmpByDate returns reservations for a emp and shift by date range
func (m *postgresDBRepo) GetReservationTypeForEmpByDate(empID int, start, end time.Time) ([]models.EmpDaysReservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.EmpDaysReservation
	query := `
		select r.id, r.start_date, r.shift_id,r.emp_id,r.user_create_id,r.wd_type_id,t.code, t.name
		from emp_days_reservations r, working_day_types t where $2 <= r.start_date and $3 >= r.start_date
		and  r.emp_id = $1 and r.wd_type_id=t.id
`

	rows, err := m.DB.QueryContext(ctx, query, empID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.EmpDaysReservation
		err := rows.Scan(
			&r.ID,
			&r.StartDate,
			&r.ShiftID,
			&r.EmpID,
			&r.UserCreateID,
			&r.WdTypeId,
			&r.WorkingDayType.Code,
			&r.WorkingDayType.Name,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) InsertReservationDayTypeForEmp(wd_type_id int, emp_id int, user_create_id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	/*query := `delete from emp_days_reservations where wd_type_id = $1 and emp_id = $2 and startDate = $3`

	_, err := m.DB.ExecContext(ctx, query, wd_type_id, emp_id, startDate)
	if err != nil {
		log.Println(err)
		return err
	}*/
	query2 := `insert into emp_days_reservations (start_date, shift_id, emp_id, user_create_id,
			created_at, updated_at, wd_type_id) values ($1, $2, $3, $4, $5, $6, $7)`
	log.Println(wd_type_id)

	var shift_id int
	if wd_type_id == 2 {
		shift_id = 2
	} else {
		shift_id = 1
	}
	if wd_type_id != 20 {
		_, err := m.DB.ExecContext(ctx, query2, startDate, shift_id, emp_id, user_create_id, time.Now(), time.Now(), wd_type_id)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// DeleteEmpDayByID deletes a emp day reservation
func (m *postgresDBRepo) DeleteEmpDayTypeByID(wd_type_id int, emp_id int, user_create_id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from emp_days_reservations where wd_type_id = $1 and emp_id = $2 and start_date = $3`

	_, err := m.DB.ExecContext(ctx, query, wd_type_id, emp_id, startDate)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
