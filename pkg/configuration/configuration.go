// Package configuration is in charge of the validation and extraction of all
// the configuration details from a configuration file or environment variables.
package configuration

import (
	"os"
	"strings"
	"time"

	"github.com/codeready-toolchain/toolchain-common/pkg/configuration"

	errs "github.com/pkg/errors"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("configuration")

var (
	// Commit current build commit set by build script.
	Commit = "0"
	// BuildTime set by build script in ISO 8601 (UTC) format:
	// YYYY-MM-DDThh:mm:ssTZD (see https://www.w3.org/TR/NOTE-datetime for
	// details).
	BuildTime = "0"
	// StartTime in ISO 8601 (UTC) format.
	StartTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
)

const (
	// EnvPrefix will be used for environment variable name prefixing.
	EnvPrefix = "REGISTRATION"

	// Constants for viper variable names. Will be used to set
	// default values as well as to get each value.
	varHTTPAddress = "http.address"
	// DefaultHTTPAddress is the address and port string that your service will
	// be exported to by default.
	DefaultHTTPAddress = "0.0.0.0:8080"

	varHTTPIdleTimeout = "http.idle_timeout"
	// DefaultHTTPIdleTimeout specifies the default timeout for HTTP idling.
	DefaultHTTPIdleTimeout = time.Second * 15

	varHTTPCompressResponses = "http.compress"
	// DefaultHTTPCompressResponses compresses HTTP responses for clients that
	// support it via the 'Accept-Encoding' header.
	DefaultHTTPCompressResponses = true

	// varEnvironment specifies service environment such as prod, stage, unit-tests, e2e-tests, dev, etc
	varEnvironment = "environment"
	// DefaultEnvironment is the default environment
	DefaultEnvironment   = "prod"
	UnitTestsEnvironment = "unit-tests"

	varLogLevel = "log.level"
	// DefaultLogLevel is the default log level used in your service.
	DefaultLogLevel = "info"

	varLogJSON = "log.json"
	// DefaultLogJSON is a switch to toggle on and off JSON log output.
	DefaultLogJSON = false

	varGracefulTimeout = "graceful_timeout"
	// DefaultGracefulTimeout is the duration for which the server gracefully
	// wait for existing connections to finish - e.g. 15s or 1m.
	DefaultGracefulTimeout = time.Second * 15

	varHTTPWriteTimeout = "http.write_timeout"
	// DefaultHTTPWriteTimeout specifies the default timeout for HTTP writes.
	DefaultHTTPWriteTimeout = time.Second * 15

	varHTTPReadTimeout = "http.read_timeout"
	// DefaultHTTPReadTimeout specifies the default timeout for HTTP reads.
	DefaultHTTPReadTimeout = time.Second * 15

	varAuthClientLibraryURL = "auth_client.library_url"
	// DefaultAuthClientLibraryURL is the default auth library location.
	DefaultAuthClientLibraryURL = "http://129.40.58.177:8082/auth/js/keycloak.js"

	varAuthClientConfigRaw     = "auth_client.config.raw"
	DefaultAuthClientConfigRaw = `{
	  "realm": "local",
	  "auth-server-url": "http://129.40.58.177:8082/auth",
	  "ssl-required": "none",
	  "resource": "crt",
	  "clientId": "crt",
	  "public-client": true
	}`

	varAuthClientConfigContentType = "auth_client.config.content_type"
	// DefaultAuthClientConfigContentType specifies the auth client config content type.
	DefaultAuthClientConfigContentType = "application/json; charset=utf-8"

	varAuthClientPublicKeysURL = "auth_client.public_keys_url"
	// DefaultAuthClientPublicKeysURL is the default log level used in your service.
	DefaultAuthClientPublicKeysURL = "http://129.40.58.177:8082/auth/realms/local/protocol/openid-connect/certs"

	varNamespace = "namespace"
	// DefaultNamespace is the default k8s namespace to use.
	DefaultNamespace = "toolchain-host-operator"

	varVerificationEnabled = "verification.enabled"
	// DefaultVerificationEnabled is the default value for whether the phone verification feature is enabled
	DefaultVerificationEnabled = false

	varVerificationDailyLimit = "verification.daily_limit"
	// DefaultVerificationDailyLimit is the default number of times a user may request phone verification
	// in a 24 hour period
	DefaultVerificationDailyLimit = 5

	varVerificationAttemptsAllowed = "verification.attempts_allowed"
	// DefaultVerificationAttemptsAllowed is the default number of maximum attempts a user may make to
	// provide a correct verification code, per generated verification code
	DefaultVerificationAttemptsAllowed = 3

	// varTwilioAccountSID is the constant used to read the configuration parameter for the
	// Twilio account identifier, used for sending SMS verification codes.  Twilio is a service that
	// provides an API for sending SMS messages anywhere in the world.  http://twilio.com
	varTwilioAccountSID = "twilio.account.sid"

	// varTwilioAuthToken is the constant used to read the configuration parameter for the
	// Twilio authentication token, used for sending SMS verification codes
	varTwilioAuthToken = "twilio.auth.token"

	// varTwilioFromNumber is the constant used to read the configuration parameter for the
	// Twilio from number, used for sending SMS verification codes
	varTwilioFromNumber = "twilio.from_number"

	// varVerificationMessageTemplate is the constant used to read the configuration parameter for the
	// verification message send to users via SMS
	varVerificationMessageTemplate = "verification.message_template"

	// DefaultVerificationMessageTemplate is the default verification message template sent to users
	// via SMS for phone verification.  The string parameter is replaced with a random verification code
	DefaultVerificationMessageTemplate = "Developer Sandbox for Red Hat OpenShift: Your verification code is %s"

	// varVerificationExcludedEmailDomains contains a comma-separated list of domain names for which phone verification
	// is not required.  For example: "redhat.com,ibm.com"
	varVerificationExcludedEmailDomains = "verification.excluded_email_domains"

	// varVerificationCodeExpiresInMin is used to set the amount of time (in minutes) that a verification code is active for before expiring
	varVerificationCodeExpiresInMin     = "verification.code_expires_in_min"
	DefaultVerificationCodeExpiresInMin = 5

	// varWoopraDomain contains the woopra domain
	varWoopraDomain = "woopra.domain"

	// varSegmentWriteKey contains the segment write key
	varSegmentWriteKey = "segment.write_key"
)

type Configuration interface {
	PrintConfig()
	GetHTTPAddress() string
	GetHTTPCompressResponses() bool
	GetHTTPWriteTimeout() time.Duration
	GetHTTPReadTimeout() time.Duration
	GetHTTPIdleTimeout() time.Duration
	GetEnvironment() string
	GetLogLevel() string
	IsLogJSON() bool
	GetGracefulTimeout() time.Duration
	IsTestingMode() bool
	GetAuthClientLibraryURL() string
	GetAuthClientConfigAuthContentType() string
	GetAuthClientConfigAuthRaw() string
	GetTwilioAccountSID() string
	GetTwilioAuthToken() string
	GetAuthClientPublicKeysURL() string
	GetNamespace() string
	GetVerificationEnabled() bool
	GetVerificationDailyLimit() int
	GetVerificationAttemptsAllowed() int
	GetVerificationMessageTemplate() string
	GetVerificationExcludedEmailDomains() []string
	GetTwilioFromNumber() string
	GetVerificationCodeExpiresInMin() int
	GetWoopraDomain() string
	GetSegmentWriteKey() string
}

// Config encapsulates the Viper configuration registry which stores the
// configuration data in-memory.
type ViperConfig struct {
	v               *viper.Viper
	secretValues    map[string]string
	excludedDomains []string
}

// LoadConfig loads the initial configuration.
func LoadConfig(cl client.Client) (*ViperConfig, error) {
	os.Setenv("HOST_OPERATOR_SECRET_NAME", "host-operator-secret")
	secret, err := configuration.LoadFromSecret("HOST_OPERATOR_SECRET_NAME", cl)
	if err != nil {
		return nil, err
	}
	os.Setenv("HOST_OPERATOR_CONFIG_MAP_NAME", "host-operator-config")
	err = configuration.LoadFromConfigMap(EnvPrefix, "HOST_OPERATOR_CONFIG_MAP_NAME", cl)
	if err != nil {
		return nil, err
	}

	return initConfig(secret), nil
}

// initConfig creates an initial, empty configuration.
func initConfig(secret map[string]string) *ViperConfig {
	c := ViperConfig{
		v:            viper.New(),
		secretValues: secret,
	}

	c.v.SetEnvPrefix(EnvPrefix)
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.setConfigDefaults()
	c.excludedDomains = strings.FieldsFunc(c.v.GetString(varVerificationExcludedEmailDomains), func(c rune) bool {
		return c == ','
	})

	return &c
}

// New creates a configuration reader object using a configurable configuration
// file path. If the provided config file path is empty, a default configuration
// will be created.
func New(configFilePath string, cl client.Client) (Configuration, error) {
	c, err := LoadConfig(cl)
	if err != nil {
		return nil, err
	}

	if configFilePath != "" {
		c.v.SetConfigType("yaml")
		c.v.SetConfigFile(configFilePath)
		err := c.v.ReadInConfig() // Find and read the config file
		if err != nil {           // Handle errors reading the config file.
			return nil, errs.Wrap(err, "failed to read config file")
		}
		if c.v.GetString(varVerificationExcludedEmailDomains) != "" {
			c.excludedDomains = strings.FieldsFunc(c.v.GetString(varVerificationExcludedEmailDomains), func(c rune) bool {
				return c == ','
			})
		}
	}
	return c, nil
}

func (c *ViperConfig) PrintConfig() {
	logWithValuesRegServ := log
	keys := c.v.AllKeys()
	for _, key := range keys {
		logWithValuesRegServ = logWithValuesRegServ.WithValues(key, c.v.Get(key))
	}

	logWithValuesRegServ.Info("Registration service configuration variables:")
}

// GetViperInstance returns the underlying Viper instance.
func (c *ViperConfig) GetViperInstance() *viper.Viper {
	return c.v
}

func (c *ViperConfig) setConfigDefaults() {
	c.v.SetTypeByDefaultValue(true)

	c.v.SetDefault(varHTTPAddress, DefaultHTTPAddress)
	c.v.SetDefault(varHTTPCompressResponses, DefaultHTTPCompressResponses)
	c.v.SetDefault(varHTTPWriteTimeout, DefaultHTTPWriteTimeout)
	c.v.SetDefault(varHTTPReadTimeout, DefaultHTTPReadTimeout)
	c.v.SetDefault(varHTTPIdleTimeout, DefaultHTTPIdleTimeout)
	c.v.SetDefault(varEnvironment, DefaultEnvironment)
	c.v.SetDefault(varLogLevel, DefaultLogLevel)
	c.v.SetDefault(varLogJSON, DefaultLogJSON)
	c.v.SetDefault(varGracefulTimeout, DefaultGracefulTimeout)
	c.v.SetDefault(varAuthClientLibraryURL, DefaultAuthClientLibraryURL)
	c.v.SetDefault(varAuthClientConfigContentType, DefaultAuthClientConfigContentType)
	c.v.SetDefault(varAuthClientPublicKeysURL, DefaultAuthClientPublicKeysURL)
	c.v.SetDefault(varNamespace, DefaultNamespace)
	c.v.SetDefault(varAuthClientConfigRaw, DefaultAuthClientConfigRaw)
	c.v.SetDefault(varVerificationEnabled, DefaultVerificationEnabled)
	c.v.SetDefault(varVerificationDailyLimit, DefaultVerificationDailyLimit)
	c.v.SetDefault(varVerificationAttemptsAllowed, DefaultVerificationAttemptsAllowed)
	c.v.SetDefault(varVerificationMessageTemplate, DefaultVerificationMessageTemplate)
	c.v.SetDefault(varVerificationCodeExpiresInMin, DefaultVerificationCodeExpiresInMin)
}

// GetHTTPAddress returns the HTTP address (as set via default, config file, or
// environment variable) that the app-server binds to (e.g. "0.0.0.0:8080").
func (c *ViperConfig) GetHTTPAddress() string {
	return c.v.GetString(varHTTPAddress)
}

// GetHTTPCompressResponses returns true if HTTP responses should be compressed
// for clients that support it via the 'Accept-Encoding' header.
func (c *ViperConfig) GetHTTPCompressResponses() bool {
	return c.v.GetBool(varHTTPCompressResponses)
}

// GetHTTPWriteTimeout returns the duration for the write timeout.
func (c *ViperConfig) GetHTTPWriteTimeout() time.Duration {
	return c.v.GetDuration(varHTTPWriteTimeout)
}

// GetHTTPReadTimeout returns the duration for the read timeout.
func (c *ViperConfig) GetHTTPReadTimeout() time.Duration {
	return c.v.GetDuration(varHTTPReadTimeout)
}

// GetHTTPIdleTimeout returns the duration for the idle timeout.
func (c *ViperConfig) GetHTTPIdleTimeout() time.Duration {
	return c.v.GetDuration(varHTTPIdleTimeout)
}

// GetEnvironment returns the environment such as prod, stage, unit-tests, e2e-tests, dev, etc
func (c *ViperConfig) GetEnvironment() string {
	return c.v.GetString(varEnvironment)
}

// GetLogLevel returns the logging level (as set via config file or environment
// variable).
func (c *ViperConfig) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}

// IsLogJSON returns if we should log json format (as set via config file or
// environment variable).
func (c *ViperConfig) IsLogJSON() bool {
	return c.v.GetBool(varLogJSON)
}

// GetGracefulTimeout returns the duration for which the server gracefully wait
// for existing connections to finish - e.g. 15s or 1m.
func (c *ViperConfig) GetGracefulTimeout() time.Duration {
	return c.v.GetDuration(varGracefulTimeout)
}

// IsTestingMode returns if the service runs in unit-tests environment
func (c *ViperConfig) IsTestingMode() bool {
	return c.GetEnvironment() == UnitTestsEnvironment
}

// GetAuthClientLibraryURL returns the auth library location (as set via
// config file or environment variable).
func (c *ViperConfig) GetAuthClientLibraryURL() string {
	return c.v.GetString(varAuthClientLibraryURL)
}

// GetAuthClientConfigAuthContentType returns the auth config config content type (as
// set via config file or environment variable).
func (c *ViperConfig) GetAuthClientConfigAuthContentType() string {
	return c.v.GetString(varAuthClientConfigContentType)
}

func (c *ViperConfig) GetAuthClientConfigAuthRaw() string {
	return c.v.GetString(varAuthClientConfigRaw)
}

// GetTwilioAccountSID is the Twilio account identifier, used for sending phone verification messages
func (c *ViperConfig) GetTwilioAccountSID() string {
	return c.secretValues[varTwilioAccountSID]
}

// GetTwilioAuthToken is the Twilio authentication token, used for sending phone verification messages
func (c *ViperConfig) GetTwilioAuthToken() string {
	return c.secretValues[varTwilioAuthToken]
}

// GetAuthClientPublicKeysURL returns the public keys URL (as set via config file
// or environment variable).
func (c *ViperConfig) GetAuthClientPublicKeysURL() string {
	return c.v.GetString(varAuthClientPublicKeysURL)
}

// GetNamespace returns the namespace in which the registration service and host operator is running
func (c *ViperConfig) GetNamespace() string {
	return c.v.GetString(varNamespace)
}

// GetVerificationEnabled indicates whether the phone verification feature is enabled or not
func (c *ViperConfig) GetVerificationEnabled() bool {
	return c.v.GetBool(varVerificationEnabled)
}

// GetVerificationDailyLimit is the number of times a user may initiate a phone verification request within a
// 24 hour period
func (c *ViperConfig) GetVerificationDailyLimit() int {
	return c.v.GetInt(varVerificationDailyLimit)
}

// GetVerificationAttemptsAllowed is the number of times a user may attempt to correctly enter a verification code,
// if they fail then they must request another code
func (c *ViperConfig) GetVerificationAttemptsAllowed() int {
	return c.v.GetInt(varVerificationAttemptsAllowed)
}

// GetVerificationMessageTemplate is the message template used to generate the content sent to users via SMS for
// phone verification
func (c *ViperConfig) GetVerificationMessageTemplate() string {
	return c.v.GetString(varVerificationMessageTemplate)
}

// GetVerificationExcludedEmailDomains returns the list of email address domains for which phone verification
// is not required
func (c *ViperConfig) GetVerificationExcludedEmailDomains() []string {
	return c.excludedDomains
}

// GetTwilioFromNumber is the phone number or alphanumeric "Sender ID" for sending phone verification messages
func (c *ViperConfig) GetTwilioFromNumber() string {
	return c.secretValues[varTwilioFromNumber]
}

// GetVerificationCodeExpiresInMin returns an int representing the number of minutes before a verification code should
// be expired
func (c *ViperConfig) GetVerificationCodeExpiresInMin() int {
	return c.v.GetInt(varVerificationCodeExpiresInMin)
}

// GetWoopraDomain returns the woopra domain name
func (c *ViperConfig) GetWoopraDomain() string {
	return c.v.GetString(varWoopraDomain)
}

// GetSegmentWriteKey returns the segment write key
func (c *ViperConfig) GetSegmentWriteKey() string {
	return c.v.GetString(varSegmentWriteKey)
}
