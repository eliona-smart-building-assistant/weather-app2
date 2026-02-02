go install github.com/aarondl/sqlboiler/v4@latest
go install github.com/aarondl/sqlboiler/v4/drivers/sqlboiler-psql@latest

go get github.com/aarondl/sqlboiler/v4
go get github.com/aarondl/null/v8

docker run --rm -d ^
    --name "eliona_database_code_generation" ^
    -e "POSTGRES_PASSWORD=secret" ^
    -p "6001:5432" ^
    -v "%cd%":/local ^
    eliona.azurecr.io/core/postgres16:latest

docker run --rm ^
    --name "eliona_database_init_code_generation" ^
    -e "CONNECTION_STRING=postgres://postgres:secret@host.docker.internal:6001/postgres" ^
    -e "INIT_CONNECTION_STRING=postgres://postgres:secret@host.docker.internal:6001/postgres" ^
    eliona.azurecr.io/core/database:tenants

docker image rm "eliona.azurecr.io/core/database:tenants"

sqlboiler psql ^
    -c sqlboiler.toml ^
    --wipe --no-tests

docker stop "eliona_database_code_generation"

go mod tidy
