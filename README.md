# 🚀 Ticket ratings
This is a gRPC service to fetch ratings of tickets - assignment for [Zendesk task](https://github.com/aa-ahmed-aa/zendesk_grpc_service/blob/master/TASK.md)

🧮 **Rating algorithm percentage:**
```
( (rating * weight) / (max_rating(5) * weight) ) * 100
```

## 🛠️ Install
You can choose one of those methods:

🐳 **Using Docker**
```bash
  cd ./zendesk_grpc_service

  docker build -t zendesk-grpc-service .

  docker run -p 50051:50051 -v $(pwd)/database.db:/app/database.db zendesk-grpc-service -n zendesk-grpc-service -rm
``` 

☸️ **For Kubernetes**
- execute this to launch the application
```bash
./deploy.sh
```
For more details on what and how this shell works check [this](https://github.com/aa-ahmed-aa/zendesk_grpc_service/blob/master/infra/K8S_SETUP.md)

- cleanup the resources 
```bash
kubectl delete -f ./infra && kubectl delete pod db-copier
```

## 🧪 Test grpc requests
Make sure you have [grpcurl](https://formulae.brew.sh/formula/grpcurl) installed

💻 **Example commands:**
```bash
cd ./zendesk_grpc_service

# 📊 CategoryScores - Spec **Aggregated category scores over a period of time**
grpcurl -plaintext \
  -import-path ./proto/ratingService/v1 \
  -proto rating_service.proto \
  -d '{"start_date":"2019-07-01","end_date":"2019-07-05"}' \
  localhost:50051 rating.RatingService/CategoryScores

# 🎫 TicketScores - Spec **Scores by ticket**
grpcurl -plaintext \
  -import-path ./proto/ratingService/v1 \
  -proto rating_service.proto \
  -d '{"start_date":"2019-07-01","end_date":"2019-07-05"}' \
  localhost:50051 rating.RatingService/TicketScores

# 🏆 OverallScore - Spec **Overall quality score**
grpcurl -plaintext \
  -import-path ./proto/ratingService/v1 \
  -proto rating_service.proto \
  -d '{"start_date":"2019-07-01","end_date":"2019-07-05"}' \
  localhost:50051 rating.RatingService/OverallScore

# 📈 ScoreChange - Spec **Period over Period score change**
grpcurl -plaintext \
  -import-path ./proto/ratingService/v1 \
  -proto rating_service.proto \
  -d '{"previous_start": "2019-07-01","previous_end": "2019-07-30","current_start": "2019-08-01","current_end": "2019-08-30"}' \
  localhost:50051 rating.RatingService/ScoreChange
```

## 📂 Folder structure
```
.
└── infra/              # ☸️ K8s resource objects
└── internal/
│   └── common/
│       ├── db.go       # 🗄️ db utiliitly
│   └── rating/
│       ├── ratingService.go
│       └── ratingRepository
├── main.go
├── proto/
│   └── ratingService/
│       └── v1/
│           ├── rating_service.proto
│           ├── rating_service.pb.go
│           └── rating_service_grpc.pb.go
```

## 📝 Commands
- 🔄 Generate the go code from the proto buff files - run this if you do any change to the `.proto` file
```bash
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/ratingService/v1/rating_service.proto
```
