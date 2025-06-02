go install golang.org/x/tools/cmd/goimports@latest

docker run --rm \
    --user $(id -u):$(id -g) \
    -v "${PWD}:/local" \
    openapitools/openapi-generator-cli:v7.13.0 generate \
    -g go-server \
    --git-user-id eliona-smart-building-assistant \
    --git-repo-id python-eliona-api-client \
    -i /local/openapi.yaml \
    -o /local/api/generated \
    --additional-properties="packageName=apiserver,sourceFolder=,outputAsLibrary=true"

goimports -w ./api/generated
