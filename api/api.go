package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/dbubel/intake"
	mw "github.com/dbubel/intake/middleware"
	"github.com/dbubel/jackstand-api/config"
	"github.com/dbubel/jackstand-api/middleware"
	"github.com/sirupsen/logrus"
)

type ServeCommand struct {
	Cfg config.Config
	Log *logrus.Logger
}

func (c *ServeCommand) Help() string {
	return "jackstand serve"
}

func (c *ServeCommand) Synopsis() string {
	return "Runs the jackstand API server"
}

func (c *ServeCommand) Run(args []string) int {
	c.Log.WithFields(logrus.Fields{"args": args}).Debug("serve command args")
	var awsConfig aws.Config

	if len(args) > 0 {
		if args[0] == "local" {
			pathStyle := true
			testEndpoint := "http://localhost:5002"
			awsConfig = aws.Config{
				Region:           aws.String("us-east-1"),
				Endpoint:         &testEndpoint,
				S3ForcePathStyle: &pathStyle,
			}
		}
	} else {
		awsConfig = aws.Config{
			Region: aws.String("us-east-1"),
		}
	}

	awsSession, err := session.NewSession(&awsConfig)

	if err != nil {
		c.Log.WithError(err).Fatalln()
	}

	//var apiKey = c.Cfg.FirebaseApiKey
	//var firebaseBaseURL = "https://www.googleapis.com/identitytoolkit/v3/relyingparty"
	//var SigninURL = fmt.Sprintf("%s/verifyPassword?key=%s", firebaseBaseURL, apiKey)
	//var CreateURL = fmt.Sprintf("%s/signupNewUser?key=%s", firebaseBaseURL, apiKey)
	//var DeleteURL = fmt.Sprintf("%s/deleteAccount?key=%s", firebaseBaseURL, apiKey)
	//var VerifyURL = fmt.Sprintf("%s/getOobConfirmationCode?key=%s", firebaseBaseURL, apiKey)
	//var ChangePasswordURL = fmt.Sprintf("%s/setAccountInfo?key=%s", firebaseBaseURL, apiKey)

	// start to create the API
	app := intake.New(c.Log)

	// Handle CORS for OPTIONS
	app.Router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	// Setup log level struct
	loglvl := mw.LogLevel{
		Log100s: true,
		Log200s: true,
		Log300s: true,
		Log400s: true,
		Log500s: true,
	}
	loggingMiddleware := mw.Logging(c.Log, loglvl)
	app.AddGlobal(loggingMiddleware)
	app.AddGlobal(mw.Recover)
	app.AddGlobal(middleware.Cors)
	app.AddGlobal(mw.Timeout(time.Second * 5))

	// Setup firebaseEndpoints struct
	fb := FireBaseAuth{
		ApiKey:          c.Cfg.FirebaseApiKey,
		FirebaseBaseURL: c.Cfg.FirebaseURL,
	}
	// Setup GetCredentialEndpoints from the firebaseEndpoints struct and apply middleware
	firebaseEndpoints := GetUserManagementEndpoints(fb)

	// Setup the Credentials struct
	creds := Credentials{
		bucket: c.Cfg.S3Bucket,
		sess:   awsSession,
		log:    c.Log,
	}
	// Setup GetCredentialEndpoints from  middleware to GetCredentialEndpoints group
	credentialEndpoints := GetCredentialEndpoints(creds, middleware.Auth)
	// Add all the GetCredentialEndpoints to the application router
	app.AddEndpoints(
		firebaseEndpoints,
		credentialEndpoints,
	)

	// Run the server
	app.Run(&http.Server{
		Addr:           fmt.Sprintf(":%d", c.Cfg.Port),
		Handler:        app.Router,
		ReadTimeout:    time.Second * 10,
		WriteTimeout:   time.Second * 10,
		MaxHeaderBytes: 1 << 20,
	})

	return 0
}
