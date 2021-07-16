package config

type Config struct {
	Port           int    `default:"4000" envconfig:"PORT"`
	S3Bucket       string `default:"jackstand-s3-test" envconfig:"S3_BUCKET"`
	LogLevel       string `default:"info" envconfig:"LOG_LEVEL"`
	JwtIssuer      string "https://securetoken.google.com/passman-fc9e0"
	JwtAud         string "passman-fc9e0"
	PublicKeyUrl   string "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
	FirebaseApiKey string `envconfig:"FIREBASE_API_KEY" required:"true"`
	firebaseURL    string `default:"https://www.googleapis.com/identitytoolkit/v3/relyingparty" envconfig:"FIREBASE_URL"`
}
