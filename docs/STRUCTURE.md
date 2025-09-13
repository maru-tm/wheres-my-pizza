restaurant-system/
├── cmd/
│   ├── kitchen/
│   │   └── run.go
│   ├── notification/
│   │   └── run.go
│   ├── order/
│   │   └── run.go
│   └── tracking/
│       └── run.go
├── config/
│   ├── config.yaml
│   └── config.go
├── docs/
│   └── ...
├── internal/
│   ├── kitchen/
│   │   ├── infrastructure/
│   │   │   ├── pg/
│   │   │   │   └── repo.go
│   │   │   └── rmq/
│   │   │       ├── consumer.go
│   │   │       ├── dto.go
│   │   │       └── publisher.go
│   │   ├── model/
│   │       ├── order.go
│   │   │   └── worker.go
│   │   └── service/
│   │       └── kitchen.go
│   ├── notification/
│   │   ├── infrastructure/
│   │   │   └── rmq/
│   │   │       ├── consumer.go
│   │   │       └── dto.go
│   │   └── service/
│   │       └── notification.go
│   ├── order/
│   │   ├── handler/
│   │       ├── dto.go
│   │   │   └── handler.go
│   │   ├── infrastructure/
│   │   │   ├── pg/
│   │   │   │   └── repo.go
│   │   │   └── rmq/
│   │   │       ├── dto.go
│   │   │       └── publisher.go
│   │   ├── model/
│   │   │   ├── validate.go
│   │   │   └── order.go
│   │   └── service/
│   │       └── order.go
│   ├── tracking/
│   │   ├── handler/
│   │   │   └── handler.go
│   │   └── service/
│   │       └── tracking.go
├── migrations/
│   ├── 001_orders.up.sql
│   ├── 002_order_items.up.sql
│   ├── 003_order_status_log.up.sql
│   ├── 004_order_sequences.up.sql
│   └── 005_workers.up.sql
├── pkg/                                
│   ├── logger/
│   │   └── logger.go
│   ├── postgres/
│   │   └── db.go
│   ├── rabbitmq/
│   │   └── rabbitmq.go
│   └── yaml/
│       └── yamlparser.go
├── docker-compose.yml
├── go.mod
├── main.go
└── README.md
