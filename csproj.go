package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Project struct {
	XMLName          xml.Name           `xml:"Project"`
	Sdk              string             `xml:"Sdk,attr"`
	Items            []PackageReference `xml:"ItemGroup>PackageReference"`
	TargetFramework  string             `xml:"PropertyGroup>TargetFramework"`
	TargetFrameworks string             `xml:"PropertyGroup>TargetFrameworks"`
}

type ItemGroup struct {
	XMLName           xml.Name           `xml:"ItemGroup"`
	PackageReferences []PackageReference `xml:"PackageReference"`
}

type PackageReference struct {
	Package    string `xml:"Include,attr"`
	Version    string `xml:"Version,attr"`
	VersionElm string `xml:"Version"`
}

func ExtractDotnetPackages(csprojPath string) (*Project, error) {
	xmlFile, err := os.Open(csprojPath)
	if err != nil {
		return nil, err
	}

	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return nil, err
	}

	var project Project

	xml.Unmarshal(byteValue, &project)

	return &project, nil
}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
