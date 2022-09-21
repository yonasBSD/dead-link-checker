package internal

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"

	"github.com/JenswBE/dead-link-checker/cmd/config"
	"github.com/JenswBE/dead-link-checker/internal/check"
	"github.com/JenswBE/dead-link-checker/internal/record"
	"github.com/JenswBE/dead-link-checker/internal/report"
	"github.com/JenswBE/dead-link-checker/reports"
)

type Manager struct {
	templates *template.Template
}

func NewManager() *Manager {
	// Parse templates
	templates := template.New("base").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	})
	templates, err := templates.ParseFS(reports.Reports, "*")
	if err != nil {
		// Panic instead of returning error as we won't be able to recover anyway.
		log.Panic().Err(err).Msg("Failed to parse build-in templates")
	}

	// Build Manager
	return &Manager{
		templates: templates,
	}
}

func (m *Manager) Run(c *config.Config) map[string]report.Report {
	// Async check all sites and generate reports from recordings
	var wg sync.WaitGroup
	reportsCollector := make(chan report.SiteReport, len(c.Sites))
	for _, siteConfig := range c.Sites {
		wg.Add(1)
		go func(siteConfig config.SiteConfig) {
			defer wg.Done()
			recorder := record.NewRecorder()
			if err := check.Run(siteConfig, recorder); err != nil {
				log.Error().Err(err).Str("site_url", siteConfig.URL.String()).Msg("Failed to run checker. Will mark as broken link.")
				recorder.RecordBrokenLink(record.BrokenLink{
					AbsoluteURL: siteConfig.URL.String(),
					BrokenLinkDetails: record.BrokenLinkDetails{
						StatusCode:        0,
						StatusDescription: "Failed to run checker: " + err.Error(),
					},
				})
			}
			recording := recorder.Stop()
			reportsCollector <- report.GenerateReport(siteConfig, recording)
		}(siteConfig)
	}

	// Wait for reports and collect in parallel
	go func() { wg.Wait(); close(reportsCollector) }()
	allReportsMap := make(map[string]report.Report, len(c.Sites))
	brokenLinksReportsMap := make(map[string]report.Report, len(c.Sites))
	for report := range reportsCollector {
		allReportsMap[report.SiteURL] = report.Report
		if len(report.BrokenLinksByPageURL) > 0 {
			brokenLinksReportsMap[report.SiteURL] = report.Report
		}
	}

	// Return if no broken links found
	if len(brokenLinksReportsMap) == 0 {
		// No broken links found
		pingHealthcheckURL(c.HealthCheck.URL)
		log.Info().Msg("No broken links found in provided sites")
		return allReportsMap
	}
	log.Info().Strs("sites", maps.Keys(brokenLinksReportsMap)).Msg("Sites with broken links found, sending notifications ...")

	// Build notifier map.
	// This map contains the notifier as key and data for this notifier as value.
	notifierMap := map[config.NotifierConfig]map[string]report.Report{}
	for _, siteConfig := range c.Sites {
		// Check if site has broken links
		siteURL := siteConfig.URL.String()
		siteReport, ok := brokenLinksReportsMap[siteURL]
		if !ok {
			// Site has no broken links => Skip site
			continue
		}

		// Add site to notifiers
		for _, notify := range siteConfig.Notify {
			notifier := c.Notifiers[notify]
			if notifierMap[notifier] != nil {
				notifierMap[notifier][siteURL] = siteReport
			} else {
				notifierMap[notifier] = map[string]report.Report{siteURL: siteReport}
			}
		}
	}

	// Send notifications
	var error_detected bool
	for notifierConfig, reportsMap := range notifierMap {
		// Generate message
		logger := log.With().Str("notifier_name", notifierConfig.NotifierName).Str("template_name", notifierConfig.TemplateName).Logger()
		logger.Debug().Msgf("Generating template for notifier '%s' ...", notifierConfig.NotifierName)
		message := &strings.Builder{}
		err := m.templates.ExecuteTemplate(message, notifierConfig.TemplateName+".html.go.tmpl", reportsMap)
		if err != nil {
			logger.Error().Err(err).Interface("reports_map", reportsMap).Msg("Failed to parse template for sending notification")
			error_detected = true
			continue
		}

		// Send message
		logger.Debug().Msgf("Sending message with notifier '%s' ...", notifierConfig.NotifierName)
		err = notifierConfig.Notifier.Send(message.String(), nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send notification")
			error_detected = true
		}
		logger.Debug().Msgf("Message sent with notifier '%s' ...", notifierConfig.NotifierName)
	}

	// Call health check
	if !error_detected {
		pingHealthcheckURL(c.HealthCheck.URL)
	}
	return allReportsMap
}

func pingHealthcheckURL(u *url.URL) {
	if u != nil {
		_, err := http.Get(u.String())
		if err != nil {
			log.Error().Err(err).Msg("Failed to send GET request to health check URL")
		}
	}
}
