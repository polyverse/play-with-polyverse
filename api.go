package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/polyverse/play-with-polyverse/config"
	"github.com/polyverse/play-with-polyverse/docker"
	"github.com/polyverse/play-with-polyverse/event"
	"github.com/polyverse/play-with-polyverse/handlers"
	"github.com/polyverse/play-with-polyverse/id"
	"github.com/polyverse/play-with-polyverse/k8s"
	"github.com/polyverse/play-with-polyverse/provisioner"
	"github.com/polyverse/play-with-polyverse/pwd"
	"github.com/polyverse/play-with-polyverse/pwd/types"
	"github.com/polyverse/play-with-polyverse/scheduler"
	"github.com/polyverse/play-with-polyverse/scheduler/task"
	"github.com/polyverse/play-with-polyverse/storage"
)

func main() {
	config.ParseFlags()

	e := initEvent()
	s := initStorage()
	df := initDockerFactory(s)

	ipf := provisioner.NewInstanceProvisionerFactory(provisioner.NewDinD(id.XIDGenerator{}, df, s))
	sp := provisioner.NewOverlaySessionProvisioner(df)

	core := pwd.NewPWD(df, e, s, sp, ipf)

	tasks := []scheduler.Task{
		task.NewCheckPorts(e, df),
		task.NewCollectStats(e, df, s),
	}
	sch, err := scheduler.NewScheduler(tasks, s, e, core)
	if err != nil {
		log.Fatal("Error initializing the scheduler: ", err)
	}

	sch.Start()

	d, err := time.ParseDuration(config.DefaultSessionDuration)
	if err != nil {
		log.Fatalf("Cannot parse duration %s. Got: %v", config.DefaultSessionDuration, err)
	}

	playground := types.Playground{
		Domain:                      config.PlaygroundDomain,
		DefaultDinDInstanceImage:    config.DefaultDinDImage,
		DefaultSessionDuration:      d,
		AvailableDinDInstanceImages: []string{config.DefaultDinDImage},
		Tasks:                       []string{".*"},
	}

	if _, err := core.PlaygroundNew(playground); err != nil {
		log.Fatalf("Cannot create default playground. Got: %v", err)
	}

	handlers.Bootstrap(core, e)
	handlers.Register(nil)
}

func initStorage() storage.StorageApi {
	if config.SessionsFile != "" {
		s, err := storage.NewFileStorage(config.SessionsFile)
		if err != nil {
			log.Fatal("Error initializing StorageAPI: ", err)
		}

		return s
	} else {
		cfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			log.Fatal("Error initializing AWS Config: ", err)
		}

		s, err := storage.S3Storage(cfg, config.S3Bucket)
		if err != nil && !os.IsNotExist(err) {
			log.Fatal("Error initializing StorageAPI: ", err)
		}
		return s
	}
}

func initEvent() event.EventApi {
	return event.NewLocalBroker()
}

func initDockerFactory(s storage.StorageApi) docker.FactoryApi {
	return docker.NewLocalCachedFactory(s)
}

func initK8sFactory(s storage.StorageApi) k8s.FactoryApi {
	return k8s.NewLocalCachedFactory(s)
}
