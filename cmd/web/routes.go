package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/tijanadmi/workinghours/internal/config"
	"github.com/tijanadmi/workinghours/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	FileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", FileServer))

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)

		mux.Get("/reservations-calendar-day-type", handlers.Repo.AdminReservationsCalendarByDayType)
		mux.Post("/reservations-calendar-day-type", handlers.Repo.AdminPostReservationsCalendarByDayType)

		mux.Get("/show-calendar", handlers.Repo.AdminShowDashboardCalendar)
		mux.Get("/show-calendar-weekly", handlers.Repo.AdminShowWeeklyDashboardCalendar)

	})

	return mux
}
