## Ticket ratings
This is a gRPC service to fetch ratings of tickets - home assignment for [Zendesk task](https://github.com/aa-ahmed-aa/zendesk_grpc_service/blob/master/TASK.md)


rating algorithm percentage = ( (rating * weight) / (max_rating(5) * weight) ) * 100

## Install
you can choose one of those methods 

Using docker 
```bash
  docker build -t zendesk-grpc-service .

  docker run -p 50051:50051 -v $(pwd)/database.db:/app/database.db zendesk-grpc-service -n zendesk-grpc-service -rm
``` 

Using kubectl 
```bash
./deploy.sh #for more details what and how this shell works check [this](https://github.com/aa-ahmed-aa/zendesk_grpc_service/blob/master/infra/README.md)
```
## folder structure
```
.
└── infra/              # K8s resource objects
└── internal/
│   └── common/
│       ├── db.go       # db utiliitly
│   └── rating/
│       ├── ratingService.go  # gRPC service implementation of business logic
│       └── ratingRepository  # the repository to execute rating sql query 
├── main.go             # Only server startup logic
├── proto/
│   └── ratingService/
│       └── v1/
│           ├── rating_service.proto
│           ├── rating_service.pb.go
│           └── rating_service_grpc.pb.go
``

## Commands
- generate the go code from the proto buff lib - run this if you do any change to the `.proto` file
```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/ratingService/v1/rating_service.proto
```