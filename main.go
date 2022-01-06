package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ebrym/bookapi/data"
	"github.com/ebrym/bookapi/handlers"
	"github.com/ebrym/bookapi/service"
	"github.com/ebrym/bookapi/utils"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/ebrym/bookapi/docs"
)

// schema for user table
const userSchema = `
		create table if not exists users (
			id 		   Varchar(36) not null,
			email 	   Varchar(100) not null unique,
			username   Varchar(225),
			password   Varchar(225) not null,
			tokenhash  Varchar(15) not null,
			isverified Boolean default false,
			createdat  Timestamp not null,
			updatedat  Timestamp not null,
			Primary Key (id)
		);
`

const verificationSchema = `
		create table if not exists verifications (
			email 		Varchar(100) not null,
			code  		Varchar(10) not null,
			expiresat 	Timestamp not null,
			type        Varchar(10) not null,
			Primary Key (email),
			Constraint fk_user_email Foreign Key(email) References users(email)
				On Delete Cascade On Update Cascade
		);
`

const categorySchema = `
		create table if not exists category (
			id 		   Varchar(36) not null,
			name 		Varchar(100) not null,
			code  		Varchar(10) not null,
			createdat  Timestamp not null,
			updatedat  Timestamp not null,
			Primary Key (id)
		);
		
`

func main() {

	logger := utils.NewLogger()

	configs := utils.NewConfigurations(logger)

	// validator contains all the methods that are need to validate the user json in request
	validator := data.NewValidation()

	// create a new connection to the postgres db store
	db, err := data.NewConnection(configs, logger)
	if err != nil {
		logger.Error("unable to connect to db", "error", err)
		panic(err)
	}
	defer db.Close()

	// creation of user table.
	db.MustExec(userSchema)
	db.MustExec(verificationSchema)
	db.MustExec(categorySchema)

	// repository contains all the methods that interact with DB to perform CURD operations for user.
	repository := data.NewPostgresRepository(db, logger)

	// repository contains all the methods that interact with DB to perform CURD operations for user.
	categoryRepository := data.NewCategoryRepository(db, logger)

	// authService contains all methods that help in authorizing a user request
	authService := service.NewAuthService(logger, configs)

	// mailService contains the utility methods to send an email
	mailService := service.NewSGMailService(logger, configs)

	// UserHandler encapsulates all the services related to user
	uh := handlers.NewAuthHandler(logger, configs, validator, repository, authService, mailService)

	// CategoryHandler encapsulates all the services related to category
	ch := handlers.NewCategoryHandler(logger, configs, validator, categoryRepository, authService)

	// create a serve mux
	sm := mux.NewRouter()

	// register handlers
	postR := sm.Methods(http.MethodPost).Subrouter()

	mailR := sm.PathPrefix("/verify").Methods(http.MethodPost).Subrouter()
	mailR.HandleFunc("/mail", uh.VerifyMail)
	mailR.HandleFunc("/password-reset", uh.VerifyPasswordReset)
	mailR.Use(uh.MiddlewareValidateVerificationData)

	postR.HandleFunc("/signup", uh.Signup)
	postR.HandleFunc("/login", uh.Login)

	// for category
	//postR.HandleFunc("/category", ch.CreateCategory)

	postR.Use(uh.MiddlewareValidateUser)

	catR := sm.Methods(http.MethodPost).Subrouter()
	catR.HandleFunc("/category", ch.CreateCategory)
	catR.Use(uh.MiddlewareValidateAccessToken)

	// used the PathPrefix as workaround for scenarios where all the
	// get requests must use the ValidateAccessToken middleware except
	// the /refresh-token request which has to use ValidateRefreshToken middleware
	refToken := sm.PathPrefix("/refresh-token").Subrouter()
	refToken.HandleFunc("", uh.RefreshToken)
	refToken.Use(uh.MiddlewareValidateRefreshToken)

	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/greet", uh.Greet)
	getR.HandleFunc("/get-password-reset-code", uh.GeneratePassResetCode)

	getR.Use(uh.MiddlewareValidateAccessToken)

	// for category
	getCat := sm.Methods(http.MethodGet).Subrouter()
	getCat.HandleFunc("/category", ch.GetCategories)
	getCat.HandleFunc("/category/{id}", ch.GetCategoryById)
	getCat.HandleFunc("/category/code/{code}", ch.GetCategoryByCode)
	getCat.Use(uh.MiddlewareValidateAccessToken)

	putR := sm.Methods(http.MethodPut).Subrouter()
	putR.HandleFunc("/update-username", uh.UpdateUsername)
	putR.HandleFunc("/reset-password", uh.ResetPassword)

	// for category
	putR.HandleFunc("/category", ch.UpdateCategory)
	putR.Use(uh.MiddlewareValidateAccessToken) // handler for documentation
	// opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	// sh := middleware.Redoc(opts, nil)

	//r := chi.NewRouter()

	sm.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:1323/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	logger.Error("Swagger plugin error : ", http.ListenAndServe(":1323", sm))
	// getD := sm.Methods(http.MethodGet).Subrouter()
	// getD.Handle("/docs", sh)
	// getD.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	// CORS
	co := gohandlers.CORS(gohandlers.AllowedOrigins([]string{"*"}))

	// create a server
	svr := http.Server{
		Addr:         configs.ServerAddress,
		Handler:      co(sm),
		ErrorLog:     logger.StandardLogger(&hclog.StandardLoggerOptions{}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// start the server
	go func() {
		logger.Info("starting the server at port", configs.ServerAddress)

		err := svr.ListenAndServe()
		if err != nil {
			logger.Error("could not start the server", "error", err)
			os.Exit(1)
		}
	}()

	// look for interrupts for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	logger.Info("shutting down the server", "received signal", sig)

	//gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	svr.Shutdown(ctx)
}
