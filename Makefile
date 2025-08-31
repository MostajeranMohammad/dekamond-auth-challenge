swagger-docs:
	swag init -g internal/application/application.go

migrate-create:
	migrate create -ext sql -dir database/migrations -seq $(name)

generate-usecase-mocks:
	~/go/bin/mockgen -source=./internal/usecases/interfaces.go -destination=./internal/usecases/mockusecases/mocks.go -package=mockusecases

generate-repository-mocks:
	~/go/bin/mockgen -source=./internal/repositories/interfaces.go -destination=./internal/repositories/mockrepositories/mocks.go -package=mockrepositories
