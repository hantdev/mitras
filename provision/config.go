package provision

import (
	"fmt"
	"os"

	"github.com/hantdev/mitras/channels"
	"github.com/hantdev/mitras/clients"
	"github.com/hantdev/mitras/pkg/errors"
	"github.com/pelletier/go-toml"
)

var errFailedToReadConfig = errors.New("failed to read config file")

// ServiceConf represents service config.
type ServiceConf struct {
	Port       string `toml:"port"          env:"MITRAS_PROVISION_HTTP_PORT"            envDefault:"9016"`
	LogLevel   string `toml:"log_level"     env:"MITRAS_PROVISION_LOG_LEVEL"            envDefault:"info"`
	TLS        bool   `toml:"tls"           env:"MITRAS_PROVISION_ENV_CLIENTS_TLS"      envDefault:"false"`
	ServerCert string `toml:"server_cert"   env:"MITRAS_PROVISION_SERVER_CERT"          envDefault:""`
	ServerKey  string `toml:"server_key"    env:"MITRAS_PROVISION_SERVER_KEY"           envDefault:""`
	ClientsURL string `toml:"clients_url"    env:"MITRAS_PROVISION_CLIENTS_LOCATION"      envDefault:"http://localhost"`
	UsersURL   string `toml:"users_url"     env:"MITRAS_PROVISION_USERS_LOCATION"       envDefault:"http://localhost"`
	HTTPPort   string `toml:"http_port"     env:"MITRAS_PROVISION_HTTP_PORT"            envDefault:"9016"`
	MgEmail    string `toml:"mitras_email"      env:"MITRAS_PROVISION_EMAIL"                envDefault:"test@example.com"`
	MgUsername string `toml:"mitras_username"   env:"MITRAS_PROVISION_USERNAME"             envDefault:"user"`
	MgPass     string `toml:"mitras_pass"       env:"MITRAS_PROVISION_PASS"                 envDefault:"test"`
	MgDomainID string `toml:"mitras_domain_id"  env:"MITRAS_PROVISION_DOMAIN_ID"            envDefault:""`
	MgAPIKey   string `toml:"mitras_api_key"    env:"MITRAS_PROVISION_API_KEY"              envDefault:""`
	MgBSURL    string `toml:"mitras_bs_url"     env:"MITRAS_PROVISION_BS_SVC_URL"           envDefault:"http://localhost:9000"`
	MgCertsURL string `toml:"mitras_certs_url"  env:"MITRAS_PROVISION_CERTS_SVC_URL"        envDefault:"http://localhost:9019"`
}

// Bootstrap represetns the Bootstrap config.
type Bootstrap struct {
	X509Provision bool                   `toml:"x509_provision" env:"MITRAS_PROVISION_X509_PROVISIONING"      envDefault:"false"`
	Provision     bool                   `toml:"provision"      env:"MITRAS_PROVISION_BS_CONFIG_PROVISIONING" envDefault:"true"`
	AutoWhiteList bool                   `toml:"autowhite_list" env:"MITRAS_PROVISION_BS_AUTO_WHITELIST"      envDefault:"true"`
	Content       map[string]interface{} `toml:"content"`
}

// Gateway represetns the Gateway config.
type Gateway struct {
	Type            string `toml:"type" json:"type"`
	ExternalID      string `toml:"external_id" json:"external_id"`
	ExternalKey     string `toml:"external_key" json:"external_key"`
	CtrlChannelID   string `toml:"ctrl_channel_id" json:"ctrl_channel_id"`
	DataChannelID   string `toml:"data_channel_id" json:"data_channel_id"`
	ExportChannelID string `toml:"export_channel_id" json:"export_channel_id"`
	CfgID           string `toml:"cfg_id" json:"cfg_id"`
}

// Cert represetns the certificate config.
type Cert struct {
	TTL string `json:"ttl" toml:"ttl" env:"MITRAS_PROVISION_CERTS_HOURS_VALID" envDefault:"2400h"`
}

// Config struct of Provision.
type Config struct {
	File          string             `toml:"file"      env:"MITRAS_PROVISION_CONFIG_FILE" envDefault:"config.toml"`
	Server        ServiceConf        `toml:"server"    mapstructure:"server"`
	Bootstrap     Bootstrap          `toml:"bootstrap" mapstructure:"bootstrap"`
	Clients       []clients.Client   `toml:"clients"   mapstructure:"clients"`
	Channels      []channels.Channel `toml:"channels"  mapstructure:"channels"`
	Cert          Cert               `toml:"cert"      mapstructure:"cert"`
	BSContent     string             `env:"MITRAS_PROVISION_BS_CONTENT" envDefault:""`
	SendTelemetry bool               `env:"MITRAS_SEND_TELEMETRY"           envDefault:"true"`
	InstanceID    string             `env:"MITRAS_MQTT_ADAPTER_INSTANCE_ID" envDefault:""`
}

// Save - store config in a file.
func Save(c Config, file string) error {
	if file == "" {
		return errors.ErrEmptyPath
	}

	b, err := toml.Marshal(c)
	if err != nil {
		return errors.Wrap(errFailedToReadConfig, err)
	}
	if err := os.WriteFile(file, b, 0o644); err != nil {
		return fmt.Errorf("Error writing toml: %w", err)
	}

	return nil
}

// Read - retrieve config from a file.
func Read(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, errors.Wrap(errFailedToReadConfig, err)
	}

	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("Error unmarshaling toml: %w", err)
	}

	return c, nil
}
