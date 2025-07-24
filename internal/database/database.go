package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jp-roisin/catch-and-go/internal/database/store"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// Sessions
	GetSession(ctx context.Context, token string) (store.Session, error)
	CreateSession(ctx context.Context, token string) (store.Session, error)
	UpdateLocale(ctx context.Context, param store.UpdateLocaleParams) error
	UpdateTheme(ctx context.Context, param store.UpdateThemeParams) error

	GetLine(ctx context.Context, param store.GetLineParams) (store.Line, error)
	ListLines(ctx context.Context) ([]store.Line, error)
	ListLinesByDirection(ctx context.Context, direction int) ([]store.Line, error)

	GetStop(ctx context.Context, code string) (store.Stop, error)

	ListStopsFromLine(ctx context.Context, id int) ([]store.Stop, error)

	CreateDashboard(ctx context.Context, param store.CreatedashboardParams) (store.Dashboard, error)
	ListDashboardsFromSession(ctx context.Context, sessionID string) ([]store.ListDashboardsFromSessionRow, error)
	DeleteDashboard(ctx context.Context, param store.DeleteDashboardParams) error
	GetDashboardById(ctx context.Context, param store.GetDashboardByIdParams) (store.Dashboard, error)
	GetDashboardByIdWithStopInfo(ctx context.Context, param store.GetDashboardByIdWithStopInfoParams) (store.GetDashboardByIdWithStopInfoRow, error)
}

type service struct {
	db      *sql.DB
	queries *store.Queries
}

var (
	dburl      = os.Getenv("BLUEPRINT_DB_URL")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	dbInstance = &service{
		db:      db,
		queries: store.New(db),
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

func (s *service) GetSession(ctx context.Context, token string) (store.Session, error) {
	return s.queries.GetSession(ctx, token)
}

func (s *service) CreateSession(ctx context.Context, token string) (store.Session, error) {
	return s.queries.CreateSession(ctx, token)
}

func (s *service) GetLine(ctx context.Context, param store.GetLineParams) (store.Line, error) {
	return s.queries.GetLine(ctx, param)
}

func (s *service) ListLines(ctx context.Context) ([]store.Line, error) {
	return s.queries.ListLines(ctx)
}
func (s *service) ListLinesByDirection(ctx context.Context, direction int) ([]store.Line, error) {
	return s.queries.ListLinesByDirection(ctx, int64(direction))
}

func (s *service) GetStop(ctx context.Context, code string) (store.Stop, error) {
	return s.queries.GetStop(ctx, code)
}

func (s *service) UpdateLocale(ctx context.Context, param store.UpdateLocaleParams) error {
	return s.queries.UpdateLocale(ctx, param)
}

func (s *service) UpdateTheme(ctx context.Context, param store.UpdateThemeParams) error {
	return s.queries.UpdateTheme(ctx, param)
}

func (s *service) ListStopsFromLine(ctx context.Context, id int) ([]store.Stop, error) {
	return s.queries.ListStopsFromLine(ctx, int64(id))
}

func (s *service) CreateDashboard(ctx context.Context, param store.CreatedashboardParams) (store.Dashboard, error) {
	return s.queries.Createdashboard(ctx, param)
}

func (s *service) ListDashboardsFromSession(ctx context.Context, sessionID string) ([]store.ListDashboardsFromSessionRow, error) {
	return s.queries.ListDashboardsFromSession(ctx, sessionID)
}

func (s *service) DeleteDashboard(ctx context.Context, param store.DeleteDashboardParams) error {
	return s.queries.DeleteDashboard(ctx, param)
}

func (s *service) GetDashboardById(ctx context.Context, param store.GetDashboardByIdParams) (store.Dashboard, error) {
	return s.queries.GetDashboardById(ctx, param)
}

func (s *service) GetDashboardByIdWithStopInfo(ctx context.Context, param store.GetDashboardByIdWithStopInfoParams) (store.GetDashboardByIdWithStopInfoRow, error) {
	return s.queries.GetDashboardByIdWithStopInfo(ctx, param)
}
