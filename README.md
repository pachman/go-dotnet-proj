# go-dotnet-proj
Parser for dotnet's project

## Run

* Add public key to [github](https://github.com/settings/keys)
* Set `SSH_PRIVATE_KEY` it will add permission to connect github 
* Set `PG_CONNECTION` => `postgres://<user>:<password>@localhost:5432/dotnet?sslmode=disable`
* (Optional) set `WARN_REPOSITORY_SIZE` if want see WARN for big repositories
* (Optional) set `FILE_MASKS` if want parse another files from ["*.csproj", "*.fsproj"]