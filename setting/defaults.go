package setting

import (
	"time"
	"bitbucket.org/tshannon/freehold/data"
	"code.google.com/p/go.crypto/bcrypt"
)

var settingDefaults map[string]Setting

func settingDefault(settingName string) *Setting {
	result, ok := settingDefaults[settingName]
	if !ok {
		panic("Default setting " + settingName + " not configured yet")
	}
	return &result
}

func init() {
	//NOTE: Settings are stored as JSON in the datastore, so all ints should be defined as floats
	// otherwise settingInt or settingFloat calls will panic
	settingDefaults = map[string]Setting{
		"LogErrors": Setting{
			Description: "Whether or not errors will be logged in the core/log datastore.",
			Value:       true,
		},
		"LogFailures": Setting{
			Description: "Whether or not input failures will be logged in the core/log datastore. " +
				"This would include items such as malformed json requests.",
			Value: false,
		},
		"LogErrorsToSyslog": Setting{
			Description: "Whether or not errors will be logged in the OS's error log.",
			Value:       false,
		},

		"DatastoreFileTimeout": Setting{
			Description: "The number of seconds of inactivity is needed for a a datastore file is " +
				"automatically closed. The higher the timeout the more resources needed to hold " +
				"open more files.  The lower the timeout the more clients will be waiting on " +
				"locked files to be opened.",
			Value:   float64(60),
			setFunc: setDatastoreTimeout,
		},
		"MinPasswordLength": Setting{
			Description: "Minimum length of new passwords.",
			Value:       float64(8),
		},
		"PasswordBcryptWorkFactor": Setting{
			Description: "Work factor used for encrypting passwords.  The higher the value the longer " +
				"it takes to encrypt a password and the safer the password is from dictionary attacks.",
			Value: float64(bcrypt.DefaultCost),
		},
		"LogFailedAuth": Setting{
			Description: "Whether or not to log failed authentication attempts.",
			Value:       true,
		},
		"LogSuccessAuth": Setting{
			Description: "Whether or not to log successful authentication attempts.",
			Value:       false,
		},
		"LogPasswordChange": Setting{
			Description: "Logs when a user changes their password.",
			Value:       false,
		},
		"404File": Setting{
			Description: "Path to a standard 404 page.",
			Value:       "/core/v1/file/404.html",
		},
		"Log404": Setting{
			Description: "Logs when an attempted is made to access an invalid resource.",
			Value:       false,
		},
		"FullClientErrors": Setting{
			Description: "If FullClientErrors is true, then the complete internal error will be " +
				"returned to the client. This can expose information about your internal system to " +
				"the public, but can be useful when troubleshooting issues, or developing applications.",
			Value: false,
		},
		"PublicRootFile": Setting{
			Description: "Path to a file that unauthenticated users get served when hitting the root of " +
				"the host: https://domain.com/.  Must be a file with public permissions.",
			Value: "/core/v1/file/public.html",
		},
		"MaxOpenSessions": Setting{
			Description: "The maximum number of open / non-expired user sessions a user can have. When a user " +
				"reaches this limit, the oldest non-expired session will be expired.",
			Value: float64(2),
		},
		"DefaultHomeApp": Setting{
			Description: "Default home app set for new users who's home app was not specified.",
			Value:       "home",
		},
		"AllowWebAppInstall": Setting{
			Description: "Whether or not applications are allowed to be installed from any arbitary url. " +
				" i.e http://github.com/developer/app/app.zip",
			Value: false,
		},
		"MaxUploadMemory": Setting{
			Description: "Maximum amount of memory (in bytes) to be used for file uploads. After max memory " +
				"is reached, temporary files are used to hold the rest of the upload.",
			Value: float64(10 * 1024 * 1024), //10MB - Minimum target raspi
		},
		"MarkdownCSSFile": Setting{
			Description: "Path to css file used for displaying markdown files.",
			Value:       "/core/v1/file/css/markdown.css",
		},
		"HttpClientTimeout": Setting{
			Description: "Timeout in seconds of requests made by the http client (like fetching an " +
				" application zip file).",
			Value: float64(15),
		},
	}
}

func setDatastoreTimeout() {
	data.SetTimeout(time.Duration(Int("DatastoreFileTimeout")) * time.Second)
}
