package main

type DotnetPackage struct {
	Package    string
	Version    string
	File       string
	Owner      string
	Repository string
}

type DotnetProjectFramework struct {
	Repository string
	File       string
	Framework  string
}

type Storage interface {
	Init(connection string) (Storage, error)
	InsertPackages(packages []DotnetPackage) error
	InsertFrameworks(frameworks []DotnetProjectFramework) error
	SelectPackages() ([]DotnetPackage, error)
	SelectRepositories() ([]string, error)
	Close()
}

func unique(intSlice []DotnetPackage) []DotnetPackage {
	keys := make(map[string]bool)
	var list []DotnetPackage
	for _, entry := range intSlice {
		key := entry.Package + entry.File
		if _, value := keys[key]; !value {
			keys[key] = true
			list = append(list, entry)
		}
	}
	return list
}

func chunk(list []DotnetPackage, chunkSize int) [][]DotnetPackage {
	var divided [][]DotnetPackage

	for i := 0; i < len(list); i += chunkSize {
		end := i + chunkSize

		if end > len(list) {
			end = len(list)
		}

		divided = append(divided, list[i:end])
	}

	return divided
}
