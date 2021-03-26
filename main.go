package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const root = "/tmp/"

var log = &Logger{}

var warnRepositorySize = 5
var fileMasks = []string{"*.csproj", "*.fsproj"}

func main() {
	envWarnRepositorySize := os.Getenv("WARN_REPOSITORY_SIZE")
	if len(envWarnRepositorySize) > 0 {
		warnRepositorySize, _ = strconv.Atoi(envWarnRepositorySize)
	}

	envFileMasks := os.Getenv("FILE_MASKS")
	if len(envFileMasks) > 0 {
		fileMasks = strings.Split(envFileMasks, "|")
	}

	connectionString := os.Getenv("PG_CONNECTION")
	if len(connectionString) < 1 {
		panic("Could you set PG_CONNECTION")
	}

	var err error
	pgsqlStorage := PgsqlStorage{}

	storage, err := pgsqlStorage.Init(connectionString)
	if err != nil {
		log.Fatal("Can't init local storage.")
		return
	}

	err = RunJob(storage)
	if err != nil {
		log.Fatalf("Can't extract packages. %v", err)
		return
	}

	defer storage.Close()

	packages, err := storage.SelectPackages()
	if err != nil {
		log.Fatalf("Can't read packages. %v", err)
		return
	}

	for _, dotnetPackage := range packages {
		println(dotnetPackage.Package)
	}
}

func RunJob(storage Storage) error {
	repositories, err := storage.SelectRepositories()
	if err != nil {
		return fmt.Errorf("RunJob | %v", err)
	}

	for _, gitUrl := range repositories {
		dotnetRepository, err := GetDotnetRepository(gitUrl)
		if err != nil {
			log.Warnf("GetDotnetRepository | %s | %v", dotnetRepository.GitUrl, err)
			continue
		}

		packages, frameworks, err := dotnetRepository.GetProjectInfo()

		if err == nil {
			if len(packages) > 0 {
				err := storage.InsertPackages(packages)
				if err != nil {
					log.Errorf("GetDotnetRepository | InsertPackages %s | %v", dotnetRepository.GitUrl, err)

					continue
				}

				log.Infof("Received packages for %s", dotnetRepository.GitUrl)
			} else {
				dotnetRepository.RemoveDirectory()
				log.Warnf("GetDotnetRepository | len(packages) == 0 | %s | %v", dotnetRepository.GitUrl, err)

				continue
			}
		} else {
			log.Errorf("GetDotnetRepository | GetProjectInfo %s | %v", dotnetRepository.GitUrl, err)
			continue
		}

		if frameworks != nil {
			err := storage.InsertFrameworks(frameworks)
			if err != nil {
				log.Errorf("GetDotnetRepository | InsertFrameworks %s | %v", dotnetRepository.GitUrl, err)

				continue
			}
		}
	}

	return nil
}

func (repository *DotnetRepository) GetProjectInfo() ([]DotnetPackage, []DotnetProjectFramework, error) {
	var dotnetPackages []DotnetPackage
	var frameworks []DotnetProjectFramework

	for _, pattern := range fileMasks {
		files, err := WalkMatch(repository.Path, pattern)
		if err != nil {
			return nil, nil, err
		}

		for _, file := range files {
			project, err := ExtractDotnetPackages(file)
			if err != nil && project.Items != nil {
				return nil, nil, err
			}

			filePath := strings.Replace(strings.ReplaceAll(file, "\\", "/"), repository.Path, "", 1)

			for _, refPk := range project.Items {

				var version = refPk.Version
				if version == "" {
					version = refPk.VersionElm
				}

				// convert X.Y.Z.W to semver X.Y.Z-W
				re := regexp.MustCompile(`^(\d+\.\d+\.\d+)\.(\d+.*)`)

				semver := re.ReplaceAllString(version, "$1-$2")

				pack := &DotnetPackage{
					Package:         refPk.Package,
					Version:         semver,
					OriginalVersion: version,
					File:            filePath,
					Owner:           repository.Project,
					Repository:      repository.Repository,
				}

				dotnetPackages = append(dotnetPackages, *pack)
			}

			if project.TargetFramework != "" {
				framework := &DotnetProjectFramework{
					Repository: repository.Repository,
					File:       filePath,
					Framework:  project.TargetFramework,
				}
				frameworks = append(frameworks, *framework)
			}

			if project.TargetFrameworks != "" {
				for _, target := range strings.Split(project.TargetFrameworks, ";") {
					framework := &DotnetProjectFramework{
						Repository: repository.Repository,
						File:       filePath,
						Framework:  target,
					}
					frameworks = append(frameworks, *framework)
				}
			}
		}
	}

	return dotnetPackages, frameworks, nil
}

//func SetupGracefulShutdown(httpServer *http.Server) {
//	signalChan := make(chan os.Signal, 1)
//	signal.Notify(
//		signalChan,
//		syscall.SIGHUP,  // kill -SIGHUP
//		syscall.SIGINT,  // kill -SIGINT or Ctrl+c
//		syscall.SIGQUIT, // kill -SIGQUIT
//	)
//
//	<-signalChan
//	log.Info("os.Interrupt - shutting down...")
//
//	go func() {
//		<-signalChan
//		log.Fatal("os.Kill - terminating...")
//	}()
//
//	gracefulCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancelShutdown()
//
//	if err := httpServer.Shutdown(gracefulCtx); err != nil {
//		log.Errorf("shutdown error: %v", err)
//		defer os.Exit(1)
//		return
//	} else {
//		log.Info("gracefully stopped")
//	}
//
//	defer os.Exit(0)
//	return
//}
