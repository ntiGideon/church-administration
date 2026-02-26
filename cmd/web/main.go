package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/joho/godotenv"
	"github.com/ntiGideon/internal/models"
)

func main() {
	gob.Register(map[string]interface{}{})

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger := slog.New(NewColoredHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":3000"
	}

	connectionString := fmt.Sprintf(
		"host=%v port=%v user=%v dbname=%s password=%s sslmode=%v",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_SSL_MODE"),
	)
	dbString := flag.String("dsn", connectionString, "database connection string")
	flag.Parse()

	db, err := connectDB(*dbString)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	dbDriver, err := OpenDriver(connectionString)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Secure = false // set true behind HTTPS
	sessionManager.Cookie.Persist = true
	sessionManager.Store = postgresstore.New(dbDriver)

	minioClient, err := newMinioClient()
	if err != nil {
		logger.Error("failed to create MinIO client", "err", err)
		os.Exit(1)
	}
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "faithconnect"
	}
	if err := ensureBucket(context.Background(), minioClient, minioBucket); err != nil {
		// Log but don't exit — MinIO might not be running in all environments.
		logger.Warn("MinIO bucket setup failed (uploads will not work)", "err", err)
	}

	app := application{
		logger:         logger,
		db:             db,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
		minioClient:       minioClient,
		minioBucket:       minioBucket,
		churchModel:       &models.ChurchModel{Db: db},
		userModel:         &models.UserModel{Db: db},
		memberModel:       &models.MemberModel{Db: db},
		eventModel:        &models.EventModel{Db: db},
		financeModel:      &models.FinanceModel{Db: db},
		invitationModel:   &models.InvitationModel{Db: db},
		announcementModel: &models.AnnouncementModel{Db: db},
		programModel:      &models.ProgramModel{Db: db},
		attendanceModel:   &models.AttendanceModel{Db: db},
		groupModel:        &models.GroupModel{Db: db},
		pledgeModel:       &models.PledgeModel{Db: db},
		rosterModel:       &models.RosterModel{Db: db},
		sermonModel:       &models.SermonModel{Db: db},
		visitorModel:      &models.VisitorModel{Db: db},
		prayerModel:       &models.PrayerRequestModel{Db: db},
		documentModel:     &models.DocumentModel{Db: db},
	}

	server := &http.Server{
		Addr:         port,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}

	logger.Info("Starting server", "addr", port)
	err = server.ListenAndServe()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
