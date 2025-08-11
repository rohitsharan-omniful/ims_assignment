package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/omniful/go_commons/config/fetcher"
	"github.com/omniful/go_commons/config/observer"
	"github.com/omniful/go_commons/constants"

	"github.com/omniful/go_commons/log"
)

type app struct {
	observer *observer.Observer
}

var (
	application *app
)

func Init(pollDuration time.Duration) (err error) {
	if getApplication() != nil {
		return
	}

	if pollDuration < time.Second*10 {
		err = errors.New("poll duration can not be less than 10 seconds")
		return
	}

	ctx := context.TODO()

	configSource, ok := os.LookupEnv("CONFIG_SOURCE")
	if !ok {
		err = errors.New("CONFIG_SOURCE env required")
	}

	if configSource != constants.LocalSource && !strings.HasPrefix(configSource, "appconfig:") {
		err = fmt.Errorf("CONFIG_SOURCE %s not valid", configSource)
		return
	}

	var pr fetcher.Fetcher

	if configSource == constants.LocalSource {
		log.Info("Reading local configuration files")

		localPath := os.Getenv(constants.ConfigPath)
		if len(localPath) == 0 {
			localPath = constants.LocalFreeFormPath
		}

		if _, err = os.Stat(localPath); os.IsNotExist(err) {
			err = fmt.Errorf("config file does not exist at path: %s", localPath)
			return
		}

		pr, err = fetcher.NewNativeFetcher(ctx, localPath)
		if err != nil {
			log.Panic("failed to initialise config: ", err)
			return
		}
	} else {
		appName := strings.TrimPrefix(configSource, "appconfig:")
		pr, err = fetcher.NewCloudFetcher(ctx, constants.RemoteFreeformProfile, appName)
		if err != nil {
			log.Panic("failed to initialise config: ", err)
			return
		}
	}

	wr, err := observer.NewObserver(ctx, pr, pollDuration)
	if err != nil {
		return
	}

	tempApp := &app{
		observer: wr,
	}

	setApplication(tempApp)
	return
}

func getApplication() *app {
	return application
}

func setApplication(app *app) {
	application = app
}
