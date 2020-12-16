module github.com/cristalhq/aconfig/aconfigenv

go 1.15

require (
	github.com/cristalhq/aconfig v0.9.1
	github.com/cristalhq/aconfig/aconfigdotenv v0.10.0-alpha
	github.com/cristalhq/aconfig/aconfigtoml v0.10.0-alpha // indirect
	github.com/joho/godotenv v1.3.0
)

replace github.com/cristalhq/aconfig => ../

replace github.com/cristalhq/aconfig/aconfigdotenv v0.10.0-alpha => ./
