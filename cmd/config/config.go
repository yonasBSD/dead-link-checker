package config

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/adhocore/gronx"
	shoutrrrRouter "github.com/containrrr/shoutrrr/pkg/router"
	shoutrrrTypes "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type RawConfig struct {
	Verbose     bool
	Cron        string
	HealthCheck RawHealthCheck `mapstructure:"health_check"`
	Notifiers   []RawNotifier  `mapstructure:"notifiers"`
	Sites       []RawSiteConfig
}

type RawHealthCheck struct {
	URL string
}

type RawNotifier struct {
	Name         string
	URL          string
	TemplateName string `mapstructure:"template_name"`
}

type RawSiteConfig struct {
	URL          string
	IgnoredLinks []string `mapstructure:"ignored_links"`
	Notify       []string
}

type Config struct {
	RawConfig
	HealthCheck HealthCheck
	Notifiers   map[string]NotifierConfig
	Sites       []SiteConfig
}

type HealthCheck struct {
	URL *url.URL
}

type NotifierConfig struct {
	Notifier     shoutrrrTypes.Sender
	NotifierName string
	TemplateName string
}

type SiteConfig struct {
	URL          *url.URL
	IgnoredLinks []*regexp.Regexp
	Notify       []string
}

// ParseConfig tries to parse the provided config path into a DeLiC config object.
func ParseConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed reading config file: %w", err)
		}
		log.Warn().Err(err).Msg("No config file found, expecting configuration through ENV variables")
	}

	// Bind ENV variables
	err = bindEnvs([]envBinding{{"Verbose", "VERBOSE"}})
	if err != nil {
		return nil, err
	}

	// Unmarshal to Config struct
	var rawConfig RawConfig
	err = viper.Unmarshal(&rawConfig, viper.DecodeHook(mapstructure.StringToSliceHookFunc(",")))
	if err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Validate and convert config
	config := Config{RawConfig: rawConfig}
	if len(rawConfig.Sites) == 0 {
		return nil, errors.New("no sites defined")
	}
	if rawConfig.Cron != "" {
		gron := gronx.New()
		if !gron.IsValid(rawConfig.Cron) {
			return nil, fmt.Errorf("invalid cron spec provided: '%s'", rawConfig.Cron)
		}
	}
	if rawConfig.HealthCheck.URL != "" {
		log.Info().Msg("No health check URL found, disabling health check calling.")
		config.HealthCheck.URL, err = url.Parse(rawConfig.HealthCheck.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse health check URL '%s': %w", rawConfig.HealthCheck.URL, err)
		}
	}
	config.Notifiers = make(map[string]NotifierConfig, len(rawConfig.Notifiers))
	serviceRouter := &shoutrrrRouter.ServiceRouter{}
	for i, rawNotifier := range rawConfig.Notifiers {
		if rawNotifier.Name == "" {
			return nil, fmt.Errorf("name missing for notifier at index %d", i)
		}
		if rawNotifier.URL == "" {
			return nil, fmt.Errorf("url missing for notifier at index %d", i)
		}
		notifierService, err := serviceRouter.Locate(rawNotifier.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse url of notifier at index %d: %s: %w", i, rawNotifier.URL, err)
		}
		config.Notifiers[rawNotifier.Name] = NotifierConfig{
			Notifier:     notifierService,
			NotifierName: rawNotifier.Name,
			TemplateName: rawNotifier.TemplateName,
		}
	}
	config.Sites = make([]SiteConfig, 0, len(rawConfig.Sites))
	for _, rawSite := range rawConfig.Sites {
		// Create initial SiteConfig
		site := SiteConfig{
			URL:          nil,
			IgnoredLinks: make([]*regexp.Regexp, 0, len(rawSite.IgnoredLinks)),
			Notify:       rawSite.Notify,
		}

		// Parse URL
		site.URL, err = url.Parse(rawSite.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse site '%s': %w", site, err)
		}

		// Parse IgnoredLinks
		for _, ignoredLink := range rawSite.IgnoredLinks {
			ignoredLinkRegex, err := regexp.Compile(ignoredLink)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ignored link '%s' for site '%s': %w", ignoredLink, site, err)
			}
			site.IgnoredLinks = append(site.IgnoredLinks, ignoredLinkRegex)
		}

		// Validate Notify
		for _, notify := range rawSite.Notify {
			if _, ok := config.Notifiers[notify]; !ok {
				return nil, fmt.Errorf("site '%s' requested to notify unknown notifier '%s'", rawSite.URL, notify)
			}
		}

		// Validation successful
		config.Sites = append(config.Sites, site)
	}
	return &config, nil
}

type envBinding struct {
	configPath string
	envName    string
}

func bindEnvs(bindings []envBinding) error {
	for _, binding := range bindings {
		err := viper.BindEnv(binding.configPath, binding.envName)
		if err != nil {
			return fmt.Errorf("failed to bind env var %s to %s: %w", binding.envName, binding.configPath, err)
		}
	}
	return nil
}
