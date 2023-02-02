package internal

import (
	"context"
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

func (m *Manager) Run(ctx context.Context, c *config.Config) map[string]report.Report {
	// Async check all sites and generate reports from recordings
	var wg sync.WaitGroup
	reportsCollector := make(chan report.SiteReport, len(c.Sites))
	for _, siteConfig := range c.Sites {
		wg.Add(1)
		go func(siteConfig config.SiteConfig) {
			defer wg.Done()
			recorder := record.NewRecorder()
			if err := check.Run(siteConfig, recorder); err != nil {
				log.Error().Err(err).Str("site_url", siteConfig.URL.String()).
					Msg("Failed to run checker. Will mark as broken link.")
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
		pingHealthCheckURL(ctx, c.HealthCheck.URL)
		log.Info().Msg("No broken links found in provided sites")
		return allReportsMap
	}
	log.Info().Strs("sites", maps.Keys(brokenLinksReportsMap)).
		Msg("Sites with broken links found, sending notifications ...")

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
	var errorDetected bool
	for notifierConfig, reportsMap := range notifierMap {
		// Generate message
		logger := log.With().
			Str("notifier_name", notifierConfig.NotifierName).
			Str("template_name", notifierConfig.TemplateName).
			Logger()
		logger.Debug().Msgf("Generating template for notifier '%s' ...", notifierConfig.NotifierName)
		message := &strings.Builder{}
		err := m.templates.ExecuteTemplate(message, notifierConfig.TemplateName+".html.go.tmpl", reportsMap)
		if err != nil {
			logger.Error().Err(err).Interface("reports_map", reportsMap).Msg("Failed to parse template for sending notification")
			errorDetected = true
			continue
		}

		// Send message
		logger.Debug().Msgf("Sending message with notifier '%s' ...", notifierConfig.NotifierName)
		err = notifierConfig.Notifier.Send(message.String(), nil)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send notification")
			errorDetected = true
		}
		logger.Debug().Msgf("Message sent with notifier '%s' ...", notifierConfig.NotifierName)
	}

	// Call health check
	if !errorDetected {
		pingHealthCheckURL(ctx, c.HealthCheck.URL)
	}
	return allReportsMap
}

func pingHealthCheckURL(ctx context.Context, u *url.URL) {
	// Skip if URL is nil
	if u == nil {
		return
	}

	// Create request
	logger := log.With().Str("health_check_url", u.String()).Logger()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create request for health check URL")
		return
	}

	// Call health check
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send GET request to health check URL")
		return
	}

	// Close body
	if resp != nil {
		if err = resp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close response body after calling health check URL")
		}
	}
}
