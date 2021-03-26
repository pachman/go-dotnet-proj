package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DotnetRepository struct {
	GitUrl     string
	Path       string
	Project    string
	Repository string
}

func GetDotnetRepository(gitUrl string) (*DotnetRepository, error) {
	dotnetRepository := GetRepositoryPath(gitUrl, root)

	var err error

	if _, err = os.Stat(dotnetRepository.Path + "/.git"); os.IsNotExist(err) {
		err = dotnetRepository.NativeGitClone()
		if err != nil {
			log.Errorf("GetDotnetRepository | NativeGitClone %s | %v", dotnetRepository.Path, err)
			dotnetRepository.RemoveDirectory()

			return nil, err
		}

		err = dotnetRepository.NativeGitSparseCheckout()
		if err != nil {
			log.Errorf("GetDotnetRepository | NativeGitSparseCheckout %s | %v", dotnetRepository.Path, err)
			dotnetRepository.RemoveDirectory()

			return nil, err
		}
	}

	err = dotnetRepository.NativeGitPull()
	if err != nil {
		dotnetRepository.RemoveDirectory()

		return nil, err
	}

	size, _ := DirSize(dotnetRepository.Path)

	if size > warnRepositorySize {
		log.Warnf("GetDotnetRepository | DirSize | Big repository [%s] size %vmb", dotnetRepository.GitUrl, size)
	}

	return dotnetRepository, nil
}

func GetRepositoryPath(gitUrl string, root string) *DotnetRepository {
	splitUrl := strings.Split(gitUrl, "/")
	repositoryName := strings.Replace(splitUrl[len(splitUrl)-1], ".git", "", -1)
	projectName := strings.Split(splitUrl[0], ":")[1]

	return &DotnetRepository{
		GitUrl:     gitUrl,
		Path:       root + projectName + "/" + repositoryName,
		Project:    projectName,
		Repository: repositoryName,
	}
}

func (repository *DotnetRepository) RemoveDirectory() {
	os.RemoveAll(repository.Path)
}

func (repository *DotnetRepository) NativeGitClone() error {
	err := os.MkdirAll(repository.Path, 0655)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repository.Path
	cmd.Run()

	cmd = exec.Command("git", "remote", "add", "origin", repository.GitUrl)
	cmd.Dir = repository.Path

	return cmd.Run()
}

func (repository *DotnetRepository) NativeGitPull() error {
	cmd := exec.Command("git", "reset", "--hard")
	cmd.Dir = repository.Path
	cmd.Run()

	cmd = exec.Command("git", "symbolic-ref", "--short", "HEAD")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Dir = repository.Path
	err := cmd.Run()
	if err != nil {
		return err
	}

	branch := strings.TrimSpace(buf.String())

	cmd = exec.Command("git", "pull", "origin", branch, "--depth=1", "--force")
	cmd.Dir = repository.Path

	return cmd.Run()
}

func (repository *DotnetRepository) NativeGitSparseCheckout() error {
	sparseMasks := ""

	for _, fileMask := range fileMasks {
		sparseMasks += "**/" + fileMask + "\n"
	}

	cmd := exec.Command("git", "sparse-checkout", "set", sparseMasks)
	cmd.Dir = repository.Path

	return cmd.Run()
}

func DirSize(path string) (int, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return int(size / (1024 * 1024)), err
}
