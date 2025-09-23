go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

go get github.com/volatiletech/sqlboiler/v4
go get github.com/volatiletech/null/v8

docker run --rm -d ^
    --name "app_sql_boiler_code_generation" ^
    -e "POSTGRES_PASSWORD=secret" ^
    -p "6001:5432" ^
    -v "%cd%":/local ^
    debezium/postgres:12

rem Wait for Postgres to initialize
timeout /t 5

rem Init
docker exec -i app_sql_boiler_code_generation ^
    psql -U postgres -f /local/db/init.sql

sqlboiler psql -c db/sqlboiler.toml --wipe --no-tests

docker stop "app_sql_boiler_code_generation"

go mod tidy
