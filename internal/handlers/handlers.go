package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tijanadmi/workinghours/internal/config"
	"github.com/tijanadmi/workinghours/internal/driver"
	"github.com/tijanadmi/workinghours/internal/forms"
	"github.com/tijanadmi/workinghours/internal/helpers"
	"github.com/tijanadmi/workinghours/internal/models"
	"github.com/tijanadmi/workinghours/internal/render"
	"github.com/tijanadmi/workinghours/internal/repository"
	"github.com/tijanadmi/workinghours/internal/repository/dbrepo"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	/*remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)*/
	m.DB.AllUsers()
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

//ShowLogin
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})

}

//PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, crudRole, gleRole, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Neispravna šifra ili lozinka!")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	if crudRole != nil {
		m.App.Session.Put(r.Context(), "crudRole", crudRole)
	}
	if gleRole != nil {
		m.App.Session.Put(r.Context(), "gleRole", gleRole)
	}
	m.App.Session.Put(r.Context(), "flash", "Uspešna prijava!")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

//Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminReservationsCalendar displays the reservation calendar
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC) //variable that contains year, month, 1-any day in that month, 0-hour, 0-minute, 0-second, 0-nanosecond, time.UTC - location
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0) // 0-number of years we wont to add, 1-number of months we wont to add, 0-number of days we wont to add
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1) // add 0-year, 1-month, -1 take one day off the current month

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	/***** Get from Session Begin****/
	user_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		//log.Println("cannot get item from session")
		m.App.ErrorLog.Println("Can't get  user_id from session")
		m.App.Session.Put(r.Context(), "error", "Can't get   from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	/***** Get from Session End****/
	employee, err := m.DB.GetEmployeeByUserIDCRUD(user_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["employee"] = employee

	for _, x := range employee {
		// create maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)
		dayblockMap := make(map[string]int)
		nightblockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
			dayblockMap[d.Format("2006-01-2")] = 0
			nightblockMap[d.Format("2006-01-2")] = 0
		}

		// get all the reservations for the current employee for day shift
		reservationsDay, err := m.DB.GetReservationForEmpByDate(1, x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range reservationsDay {

			// it's a block
			blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			dayblockMap[y.StartDate.Format("2006-01-2")] = y.ID
			fmt.Printf("EMP_ID=%d, Blocked value=%d For day=%s ", x.ID, dayblockMap[y.StartDate.Format("2006-01-2")], y.StartDate.Format("2006-01-2"))
			fmt.Println()

		}

		// get all the reservations for the current employee for day shift
		reservationsNight, err := m.DB.GetReservationForEmpByDate(2, x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range reservationsNight {

			// it's a block

			nightblockMap[y.StartDate.Format("2006-01-2")] = y.ID
			fmt.Printf("EMP_ID=%d, Blocked value=%d For day=%s ", x.ID, nightblockMap[y.StartDate.Format("2006-01-2")], y.StartDate.Format("2006-01-2"))
			fmt.Println()

		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap
		data[fmt.Sprintf("day_block_map_%d", x.ID)] = dayblockMap
		data[fmt.Sprintf("night_block_map_%d", x.ID)] = nightblockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
		m.App.Session.Put(r.Context(), fmt.Sprintf("day_block_map_%d", x.ID), dayblockMap)
		m.App.Session.Put(r.Context(), fmt.Sprintf("night_block_map_%d", x.ID), nightblockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminPostReservationsCalendar handles post of reservation calendar
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	/***** Get from Session Begin****/
	user_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		//log.Println("cannot get item from session")
		m.App.ErrorLog.Println("Can't get  user_id from session")
		m.App.Session.Put(r.Context(), "error", "Can't get   from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	/***** Get from Session End****/
	// process blocks
	employee, err := m.DB.GetEmployeeByUserIDCRUD(user_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, x := range employee {
		// Get the block map from the session. Loop through entire map, if we have an entry in the map
		// that does not exist in our posted data, and if the restriction id > 0, then it is a block we need to
		// remove.
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("day_block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_dayblock_%d_%s", x.ID, name), r) {
						// delete the restriction by id
						err := m.DB.DeleteEmpDayByID(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	for _, x := range employee {
		// Get the block map from the session. Loop through entire map, if we have an entry in the map
		// that does not exist in our posted data, and if the restriction id > 0, then it is a block we need to
		// remove.
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("night_block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_nightblock_%d_%s", x.ID, name), r) {
						// delete the restriction by id
						err := m.DB.DeleteEmpDayByID(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	// now handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_dayblock") {
			exploded := strings.Split(name, "_")
			employeeID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			// insert a new block
			err := m.DB.InsertReservationDayForEmp(1, employeeID, user_id, t)
			if err != nil {
				log.Println(err)
			}
		}
	}

	// now handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_nightblock") {
			exploded := strings.Split(name, "_")
			employeeID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			// insert a new block
			err := m.DB.InsertReservationDayForEmp(2, employeeID, user_id, t)
			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Izmene su sačuvane!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

// AdminReservationsCalendar displays the reservation calendar
func (m *Repository) AdminShowDashboardCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC) //variable that contains year, month, 1-any day in that month, 0-hour, 0-minute, 0-second, 0-nanosecond, time.UTC - location
	}

	/***** Get from Session Begin****/
	user_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		//log.Println("cannot get item from session")
		m.App.ErrorLog.Println("Can't get  user_id from session")
		m.App.Session.Put(r.Context(), "error", "Can't get   from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	/***** Get from Session End****/

	/**** get OrgUnits for the list Begin***/

	orgUnitsList, err := m.DB.GetOrgUnitsByUserIDGLE(user_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	/**** get OrgUnits for the list End***/
	var org_id int
	if r.URL.Query().Get("o") != "" {
		org_id, _ = strconv.Atoi(r.URL.Query().Get("o"))
	} else {
		org_id = orgUnitsList[0].ID
	}

	data := make(map[string]interface{})
	data["now"] = now
	data["orgUnitsList"] = orgUnitsList
	next := now.AddDate(0, 1, 0) // 0-number of years we wont to add, 1-number of months we wont to add, 0-number of days we wont to add
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["org_id"] = strconv.Itoa(org_id)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1) // add 0-year, 1-month, -1 take one day off the current month

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	employee, err := m.DB.GetEmployeeByOrgID(org_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["employee"] = employee

	for _, x := range employee {
		// create maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)
		dayblockMap := make(map[string]int)
		nightblockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
			dayblockMap[d.Format("2006-01-2")] = 0
			nightblockMap[d.Format("2006-01-2")] = 0
		}

		// get all the reservations for the current employee for day shift
		reservationsDay, err := m.DB.GetReservationForEmpByDate(1, x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range reservationsDay {

			// it's a block
			blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			dayblockMap[y.StartDate.Format("2006-01-2")] = y.ID
			fmt.Printf("EMP_ID=%d, Blocked value=%d For day=%s ", x.ID, dayblockMap[y.StartDate.Format("2006-01-2")], y.StartDate.Format("2006-01-2"))
			fmt.Println()

		}

		// get all the reservations for the current employee for day shift
		reservationsNight, err := m.DB.GetReservationForEmpByDate(2, x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range reservationsNight {

			// it's a block

			nightblockMap[y.StartDate.Format("2006-01-2")] = y.ID
			fmt.Printf("EMP_ID=%d, Blocked value=%d For day=%s ", x.ID, nightblockMap[y.StartDate.Format("2006-01-2")], y.StartDate.Format("2006-01-2"))
			fmt.Println()

		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap
		data[fmt.Sprintf("day_block_map_%d", x.ID)] = dayblockMap
		data[fmt.Sprintf("night_block_map_%d", x.ID)] = nightblockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
		m.App.Session.Put(r.Context(), fmt.Sprintf("day_block_map_%d", x.ID), dayblockMap)
		m.App.Session.Put(r.Context(), fmt.Sprintf("night_block_map_%d", x.ID), nightblockMap)
	}

	render.Template(w, r, "admin-show-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminReservationsCalendar displays the reservation calendar
func (m *Repository) AdminShowWeeklyDashboardCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC) //variable that contains year, month, 1-any day in that month, 0-hour, 0-minute, 0-second, 0-nanosecond, time.UTC - location
	}
	now.AddDate(0, 0, -now.Day()+1)
	weekday := int(now.AddDate(0, 0, -now.Day()+1).Weekday())
	fmt.Println(weekday) // "Tuesday"

	/***** Get from Session Begin****/
	user_id, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		//log.Println("cannot get item from session")
		m.App.ErrorLog.Println("Can't get  user_id from session")
		m.App.Session.Put(r.Context(), "error", "Can't get   from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	/***** Get from Session End****/

	/**** get OrgUnits for the list Begin***/

	orgUnitsList, err := m.DB.GetOrgUnitsByUserIDGLE(user_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	/**** get OrgUnits for the list End***/
	var org_id int
	if r.URL.Query().Get("o") != "" {
		org_id, _ = strconv.Atoi(r.URL.Query().Get("o"))
	} else {
		org_id = orgUnitsList[0].ID
	}

	data := make(map[string]interface{})
	data["now"] = now
	data["orgUnitsList"] = orgUnitsList
	next := now.AddDate(0, 1, 0) // 0-number of years we wont to add, 1-number of months we wont to add, 0-number of days we wont to add
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["org_id"] = strconv.Itoa(org_id)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1) // add 0-year, 1-month, -1 take one day off the current month

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()
	intMap["weekday"] = weekday

	for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
		employee, err := m.DB.GetReservationEmployeeByDate(1, org_id, d)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		data[fmt.Sprintf("day_block_map_%d", d.Day())] = employee

		employee, err = m.DB.GetReservationEmployeeByDate(2, org_id, d)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		data[fmt.Sprintf("night_block_map_%d", d.Day())] = employee
	}

	render.Template(w, r, "admin-show-weekly-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}
