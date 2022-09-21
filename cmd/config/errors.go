package config

import (
	"errors"
	"fmt"
)

var ErrNoSitesDefined = errors.New("no sites defined")

type InvalidCronSpecError struct {
	spec string
}

func (e InvalidCronSpecError) Error() string {
	return fmt.Sprintf("invalid cron spec '%s' provided", e.spec)
}

type NotifierFieldMissingError struct {
	field string
	index int
}

func (e NotifierFieldMissingError) Error() string {
	return fmt.Sprintf("%s missing for notifier at index %d", e.field, e.index)
}

type UnknownNotifierForSiteError struct {
	siteURL  string
	notifier string
}

func (e UnknownNotifierForSiteError) Error() string {
	return fmt.Sprintf("site '%s' requested to notify unknown notifier '%s'", e.siteURL, e.notifier)
}
