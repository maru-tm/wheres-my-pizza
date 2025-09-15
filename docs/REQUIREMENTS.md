# wheres-my-pizza

## General Criteria

- Your code MUST be written in accordance with [gofumpt](https://github.com/mvdan/gofumpt). If not, you will automatically receive a score of `0`.
- Your program MUST compile successfully.
- Your program MUST NOT crash unexpectedly (any panics: `nil-pointer dereference`, `index out of range`, etc.). If this happens, you will receive `0` points during defense.
- Only built-in Go packages, `pgx/v5` and the official AMQP client (`github.com/rabbitmq/amqp091-go`) are allowed. If other packages are used, you will receive a score of `0`.
- RabbitMQ server MUST be running and available for connection.
- PostgreSQL database MUST be running and accessible for all services
- All RabbitMQ connections must handle reconnection scenarios
- Implement proper graceful shutdown for all services
- All database operations must be transactional where appropriate
- The project MUST compile with the following command in the project root directory:

```sh
$ go build -o restaurant-system .
```

### Logging Format

All services must implement structured logging in JSON format and write to `stdout`.

- **Core Fields (Mandatory):** `timestamp`, `level`, `service`, `action`, `message`, `hostname`, `request_id`.
- **Error Object:** For `ERROR` level logs, an `error` object with `msg`, and `stack` must be included.

#### Field Descriptions

| Key          | Type   | Description                                                                                |
| :----------- | :----- | :----------------------------------------------------------------------------------------- |
| `timestamp`  | string | Time the log entry was created, in ISO 8601 format.                                        |
| `level`      | string | Severity level of the event. Valid values: `INFO`, `DEBUG`, `ERROR`.                       |
| `service`    | string | Name of the service that generated the log (e.g., `order-service`, `kitchen-worker`).      |
| `action`     | string | Short, machine-readable event name (e.g., `order_created`, `db_error`, `retry_attempted`). |
| `message`    | string | Human-readable message describing the event.                                               |
| `hostname`   | string | Hostname or container ID of the service emitting the log.                                  |
| `request_id` | string | Correlation ID for tracing a single operation across multiple services.                    |
| `error`      | object | _(Only for `ERROR` logs)_ Contains structured error information.                           |
| ├─ `msg`     | string | Error message explaining the cause.                                                        |
| └─ `stack`   | string | Stack trace or error traceback for debugging purposes.                                     |

### Configuration

Services should be configurable via a YAML file (`config.yaml`) for database and RabbitMQ connection details.

```yaml
# Database Configuration
database:
  host: localhost
  port: 5432
  user: restaurant_user
  password: restaurant_pass
  database: restaurant_db

# RabbitMQ Configuration
rabbitmq:
  host: localhost
  port: 5672
  user: guest
  password: guest
```

## System Architecture Overview

Your application will consist of four main services, a database, and a message broker. They interact as follows:

```
                                +--------------------------------------------+
                                |               PostgreSQL DB                |
                                |             (Order Storage)                |
                                +--+-------------+---------------------------+
                                   ^             ^                    |
                  (Writes & Reads) |             | (Writes & Reads)   |
                                   v             v                    |
+------------+        +-----------+              +---------------+    |
| HTTP Client|------->|  Order    |              | Kitchen       |    |
| (e.g. curl)|        |  Service  |              | Service       |    |
+------------+        +---------- +              +-+-------------+    |
                         |                         ^                  |
                (Publishes New Order)    (Publishes Status Update)    |
                         v                         |                  |
                   +-----+-------------------------+---------+        |
                   |                                         |        |
                   |         RabbitMQ Message Broker         |        |
                   |                                         |        |
                   +-----------------------------------------+        |
                              |                                       |
                              | (Status Updates)                      | (Reads)
                              v                                       v
                        +-----+-----------+         +-----+------------------+
                        | Notification    |         | Tracking               |
                        | Subscriber      |         | Service                |
                        +-----------------+         +------------------------+
```

## Feature: Order Service

### Context

The Order Service is the public-facing entry point of the restaurant system. Its primary responsibility is to receive new orders from customers via an HTTP API, validate them, store them in the database, and publish them to a message queue for the kitchen staff to process. It acts as the gatekeeper, ensuring all incoming data is correct and formatted before entering the system.

### Migrations

This service interacts with the following database tables. It is responsible for inserting new records into them.

- **`orders` Table**: The primary table for storing all order details.

- **Possible statuses for `orders`:**

| Status      | Description                                         |
| ----------- | --------------------------------------------------- |
| `received`  | Order has been received and is awaiting processing. |
| `cooking`   | Order is currently being prepared in the kitchen.   |
| `ready`     | Order is ready for pickup or delivery.              |
| `completed` | Order has been successfully delivered or picked up. |
| `cancelled` | Order has been cancelled.                           |

```sql
create table "orders" (
    "id"                serial        primary key,
    "created_at"        timestamptz   not null    default now(),
    "updated_at"        timestamptz   not null    default now(),
    "number"            text          unique not null,
    "customer_name"     text          not null,
    "type"              text          not null check (type in ('dine_in', 'takeout', 'delivery')),
    "table_number"      integer,
    "delivery_address"  text,
    "total_amount"      decimal(10,2) not null,
    "priority"          integer       default 1,
    "status"            text          default 'received',
    "processed_by"      text,
    "completed_at"      timestamptz
);
```

- **`order_items` Table**: Stores the individual items associated with each order.

```sql
create table order_items (
    "id"          serial        primary key,
    "created_at"  timestamptz   not null    default now(),
    "order_id"    integer       references orders(id),
    "name"        text          not null,
    "quantity"    integer       not null,
    "price"       decimal(8,2)  not null
);
```

- **`order_status_log` Table**: Creates an audit trail for an order's lifecycle, starting with the `received` status.

- **Possible statuses for `order_status_log`:**

| Status      | Description                                             |
| ----------- | ------------------------------------------------------- |
| `received`  | The order has been received and is awaiting processing. |
| `cooking`   | The order is currently being prepared in the kitchen.   |
| `ready`     | The order is ready for pickup or delivery.              |
| `completed` | The order has been successfully delivered or picked up. |
| `cancelled` | The order has been cancelled.                           |

```sql
create table order_status_log (
    "id"          serial        primary key,
    "created_at"  timestamptz   not null    default now(),
    "order_id"    integer       references orders(id),
    "status"      text,
    "changed_by"  text,
    "changed_at"  timestamptz   default current_timestamp,
    "notes"       text
);
```

### API

#### Create new customer order

```http
POST /orders
```

#### Request Body (`application/json`)

```json
{
  "customer_name": "John Doe",
  "order_type": "takeout",
  "items": [
    { "name": "Margherita Pizza", "quantity": 1, "price": 15.99 },
    { "name": "Caesar Salad", "quantity": 1, "price": 8.99 }
  ]
}
```

#### Response Body (`application/json`)

```json
{
  "order_number": "ORD_20241216_001",
  "status": "received",
  "total_amount": 24.98
}
```

### Logic

1. **Receive Request:** The service listens for an HTTP `POST` request on the `/orders` endpoint.

2. **Input Validation:** The incoming JSON payload is validated against the following rules:

| `tag`           | `required type` | `description`                                                                                      |
| --------------- | --------------- | -------------------------------------------------------------------------------------------------- |
| `customer_name` | strin           | 1-100 characters. Must not contain special characters other than spaces, hyphens, and apostrophes. |
| `order_type`    | strin           | Must be one of: `'dine_in'`, `'takeout'`, or `'delivery'`.                                         |
| `items`         | arra            | Must contain between 1 and 20 items.                                                               |
| `item.name`     | strin           | 1-50 characters.                                                                                   |
| `item.quantity` | intege          | must be between 1 and 10.                                                                          |
| `item.price`    | decima          | must be between `0.01` and `999.99`.                                                               |

- **Conditional Validation:**

| `order_type` | `required fields`                              | `description`                                 | `must not be present` |
| ------------ | ---------------------------------------------- | --------------------------------------------- | --------------------- |
| `'dine_in'`  | `table_number` (integer, 1-100)                | Table number at which the customer is served. | `delivery_address`    |
| `'delivery'` | `delivery_address` (string, min 10 characters) | Address for delivery of the order by courier. | `table_number`        |

3. **Data Processing:**

    - **Calculate `total_amount`:** Sum the `price` \* `quantity` for all items in the order.
    - **Assign `priority`:** Determine the order's priority based on its total amount. The highest applicable priority is used.

      | Priority | Criteria                                    |
           | -------- | ------------------------------------------- |
      | `10`     | Order total amount is greater than $100.    |
      | `5`      | Order total amount is between $50 and $100. |
      | `1`      | All other standard orders.                  |

    - **Generate `order_number`:** Create a unique order number using the format `ORD_YYYYMMDD_NNN`. The `NNN` sequence should reset to `001` daily (based on UTC). This must be managed transactionally to prevent race conditions.

4. **Database Transaction:** All database operations must be executed within a single transaction to ensure data integrity.

    - Insert the main order details into the `orders` table.
    - Insert each item from the `items` array into the `order_items` table, linked by the new `order_id`.
    - Log the initial status by inserting a record into the `order_status_log` table with `status: 'received'`.

5. **Message Publishing:**

    - After the transaction is successfully committed, publish an `Order Message` to the `orders_topic` exchange in RabbitMQ.
    - **Routing Key:** `kitchen.{order_type}.{priority}` (e.g., `kitchen.takeout.1`).
    - **Message Properties:** The message must be persistent (`delivery_mode: 2`) and have its `priority` property set according to the assigned priority level.
    - **Message Format:**

   ```json
   {
     "order_number": "ORD_20241216_001",
     "customer_name": "John Doe",
     "order_type": "delivery",
     "table_number": null,
     "delivery_address": "123 Main St, City",
     "items": [{ "name": "Margherita Pizza", "quantity": 7, "price": 15.99 }],
     "total_amount": 111.93,
     "priority": 10
   }
   ```

6. **Error Handling:** If validation fails, return an HTTP `400 Bad Request` response with a clear error message. If a database or RabbitMQ error occurs, return an HTTP `500 Internal Server Error`.

7. **Send Response:** Return an HTTP `200 OK` response with the generated `order_number`, `status`, and `total_amount`.

### Logs

All log entries must be structured JSON written to `stdout`. Refer to the global [Logging Format](#logging-format) section for core fields (`timestamp`, `level`, `service`, etc.).

| Level | Action                  | Description                                                                         |
| ----- | ----------------------- | ----------------------------------------------------------------------------------- |
| INFO  | service_started         | On successful startup.                                                              |
| INFO  | db_connected            | On successful database connection.                                                  |
| INFO  | rabbitmq_connected      | On successful RabbitMQ connection.                                                  |
| DEBUG | order_received          | When a new valid order is received.                                                 |
| DEBUG | order_published         | When an order message is successfully published to RabbitMQ.                        |
| ERROR | validation_failed       | When incoming order data fails validation. Include details of the validation error. |
| ERROR | db_transaction_failed   | If the database transaction for creating an order fails.                            |
| ERROR | rabbitmq_publish_failed | If publishing the order message fails.                                              |

### Flags

| Option             | Type   | Default | Description                                     |
| ------------------ | ------ | ------- | ----------------------------------------------- |
| `--mode`           | string | —       | **Required.** Must be set to `order-service`.   |
| `--port`           | int    | `3000`  | The HTTP port for the API.                      |
| `--max-concurrent` | int    | `50`    | Maximum number of concurrent orders to process. |

### Usage

#### Launch the service

```sh
./restaurant-system --mode=order-service --port=3000
```

_Expected Log Output:_

```json
{"timestamp":"2024-12-16T10:30:00.000Z","level":"INFO","service":"order-service","hostname":"order-service-789abc","request_id":"startup-001","action":"service_started","message":"Order Service started on port 3000","details":{"port":3000,"max_concurrent":50}}
{"timestamp":"2024-12-16T10:30:01.000Z","level":"INFO","service":"order-service","hostname":"order-service-789abc","request_id":"startup-001","action":"db_connected","message":"Connected to PostgreSQL database","duration_ms":250}
{"timestamp":"2024-12-16T10:30:02.000Z","level":"INFO","service":"order-service","hostname":"order-service-789abc","request_id":"startup-001","action":"rabbitmq_connected","message":"Connected to RabbitMQ exchange 'orders_topic'","duration_ms":150}
```

#### Place an order using `curl`

```sh
curl -X POST http://localhost:3000/orders \
  -H "Content-Type: application/json" \
  -d '{
        "customer_name": "Jane Doe",
        "order_type": "takeout",
        "items": [
          {"name": "Margherita Pizza", "quantity": 1, "price": 15.99},
          {"name": "Caesar Salad", "quantity": 1, "price": 8.99}
        ]
      }'
```

---

## Feature: Kitchen Worker

### Context

The Kitchen Worker is a background service that simulates the kitchen staff. It consumes order messages from a queue, processes them, and updates their status in the database. It is the core processing engine of the restaurant. Multiple worker instances can run concurrently to handle high order volumes and can be specialized to process specific types of orders.

### Migrations

This service interacts with the following database tables. It is responsible for creating worker records and updating order statuses.

- **`workers` Table**: A registry for all kitchen workers, their specializations, and their current status.

  ```sql
  create table workers (
      "id"                serial      primary key,
      "created_at"        timestamptz not null    default now(),
      "name"              text        unique not null,
      "type"              text        not null,
      "status"            text        default 'online',
      "last_seen"         timestamptz default current_timestamp,
      "orders_processed"  integer     default 0
  );
  ```

- **`orders` Table**: References this table to update order `status`, `completed_at`, and `processed_by` fields. (See [Order Service Migrations](#migrations) for table definition).
- **`order_status_log` Table**: References this table to insert status change events (`cooking`, `ready`). (See [Order Service Migrations](#migrations) for table definition).

### API

This service does not expose any API endpoints. It interacts with the system via RabbitMQ messages.

### Logic

1. **Startup and Registration:**

- On startup, the worker must register itself by inserting (or updating) a record in the `workers` table with its unique name and type, marking itself **online**.
- **Duplicate check:**

    - If a record with the same `worker_name` already exists **and is marked online**, the worker must log `ERROR: worker_registration_failed` and terminate with exit code `1`.
    - If a record with the same name exists **but is marked offline**, the new worker may safely start: it should update the existing record to reflect its set the status to **online**.

2. **Message Consumption:**

- The worker connects to RabbitMQ and consumes messages from one or more kitchen queues (`kitchen_queue`, `kitchen_dine_in_queue`, etc.) which are bound to the `orders_topic` exchange.
- It must configure a prefetch count (`basic.qos`) to limit the number of unacknowledged messages it holds at once.

  > In RabbitMQ, `basic.qos` is a method used to configure the quality of service for consumers, specifically by controlling how many messages a consumer can receive without acknowledging them.
  >
  > _Using `basic.qos` is critical to prevent worker overload and distribute the load evenly between workers._

3. **Worker Specialization:**

- If the `--order-types` flag is specified, the worker checks the `order_type` of the incoming message.
- If the worker is not specialized for that order type, it must negatively acknowledge the message (`basic.nack`) with `requeue=true`, returning it to the queue for another worker.

  > _Using `basic.nack` with `requeue=true` is a way to return a message to the queue if the worker is not intended to handle it._

4. **Order Processing:**

- **Set Status to `cooking`:** Update the order's status to `'cooking'` and set `processed_by` to the worker's name in the `orders` table.
- **Log Status Change:** Insert a new record in `order_status_log`.
- **Transactional integrity**: Changes to the database (such as updating status or logging) must be transactional. If any step (e.g., inserting into order_status_log) fails, the previous status update in orders must be rolled back.
- **Idempotency**: If a message is reprocessed (due to a `nack` or failure), the worker must be able to detect that the order is already in the cooking status and avoid setting it again, in order to prevent redundant log entries or notifications.
- **Publish Status Update:** Publish a `Status Update Message` to the `notifications_fanout` exchange.

    - **Message Format:**

      ```json
      {
        "order_number": "ORD_20241216_001",
        "old_status": "received",
        "new_status": "cooking",
        "changed_by": "chef_mario",
        "timestamp": "2024-12-16T10:32:00Z",
        "estimated_completion": "2024-12-16T10:42:00Z"
      }
      ```

- **Simulate Cooking:** Pause execution to simulate cooking time based on the `order_type`:
    - `dine_in`: 8 seconds.
    - `takeout`: 10 seconds.
    - `delivery`: 12 seconds.
- **Set Status to `ready`:** Update the order's status to `'ready'` and set the `completed_at` timestamp in the `orders` table. Increment the worker's `orders_processed` count in the `workers` table.
- **Log and Publish Again:** Log the status change to `order_status_log` and publish another `Status Update Message` to the `notifications_fanout` exchange.

5. **Message Acknowledgement:** Once all processing steps are complete, send a positive acknowledgement (`basic.ack`) to RabbitMQ to permanently remove the message from the queue.

   > _In RabbitMQ, `basic.ack` is a method used for positive acknowledgments. When a consumer receives a message and processes it successfully, it should send a basic.ack to the broker (RabbitMQ) to indicate that the message has been handled and can be removed from the queue. This ensures that messages are not lost if a consumer crashes or disconnects, as they will be redelivered to another consumer._

6. **Heartbeat Mechanism:**

- The worker sends a periodic heartbeat by updating its `last_seen` timestamp and `status` to `'online'` in the `workers` table.

7. **Error and Redelivery Handling:**

- If any processing step fails (e.g., database unavailable), the worker must negatively acknowledge the message (`basic.nack`) with `requeue=true` so it can be re-processed later.
- If there are data validation errors in the message that will never be corrected, the message should be sent to a Dead-Letter Queue (`DLQ`) to allow for manual analysis and to prevent the queue from being blocked.

  > Dead Letter Queue (DLQ) is a specialized queue that stores messages that cannot be delivered or processed by their intended queue. It acts as a safety net, preventing failed messages from being lost and allowing for inspection, troubleshooting, and potential reprocessing.

8. **Graceful Shutdown:**
    - On receiving `SIGINT` or `SIGTERM`, the worker must:
        1. Stop consuming new messages from the queue.
        2. Finish processing its current in-flight order.
        3. Update its status to `'offline'` in the `workers` table.
        4. `nack` any other unacknowledged messages to requeue them.
        5. Exit gracefully.

### Logs

| Level | Action                     | Description                                         |
| ----- | -------------------------- | --------------------------------------------------- |
| INFO  | worker_registered          | On successful registration in the `workers` table.  |
| DEBUG | order_processing_started   | When an order is picked from the queue.             |
| DEBUG | order_completed            | When an order is fully processed.                   |
| DEBUG | heartbeat_sent             | When a heartbeat is successfully sent.              |
| INFO  | graceful_shutdown          | When the worker starts its shutdown sequence.       |
| ERROR | worker_registration_failed | If the worker name is a duplicate.                  |
| ERROR | message_processing_failed  | For unrecoverable processing errors.                |
| ERROR | db_connection_failed       | When the database is unreachable after all retries. |

### Flags

| Option                 | Type   | Default | Description                                                                                                             |
| ---------------------- | ------ | ------- | ----------------------------------------------------------------------------------------------------------------------- |
| `--mode`               | string | —       | **Required.** Must be set to `kitchen-worker`.                                                                          |
| `--worker-name`        | string | —       | **Required.** Unique name for the worker (e.g., `chef_mario`).                                                          |
| `--order-types`        | string | —       | Optional. Comma-separated list of order types the worker can handle (e.g., `dine_in,takeout`). If omitted, handles all. |
| `--heartbeat-interval` | int    | `30`    | Interval (seconds) between heartbeats.                                                                                  |
| `--prefetch`           | int    | `1`     | RabbitMQ prefetch count, limiting how many messages the worker receives at once.                                        |

### Usage

1. **Launch a general worker:**

   ```sh
   ./restaurant-system --mode=kitchen-worker --worker-name="chef_anna" --prefetch=1
   ```

2. **Launch specialized workers:**

   ```sh
   # This worker only handles dine-in orders
   ./restaurant-system --mode=kitchen-worker --worker-name="chef_mario" --order-types="dine_in" &

   # This worker only handles delivery orders
   ./restaurant-system --mode=kitchen-worker --worker-name="chef_luigi" --order-types="delivery" &
   ```

---

## Feature: Tracking Service

### Context

The Tracking Service provides visibility into the restaurant's operations. It offers a read-only HTTP API for external clients (like a customer-facing app or an internal dashboard) to query the current status of orders, view an order's history, and monitor the status of all kitchen workers. It directly queries the database and does not interact with RabbitMQ.

### Migrations

This service reads data from the following tables but does not write to them.

- **`orders` Table**: Referenced to get the current status of an order. (See [Order Service Migrations](#migrations) for definition).
- **`order_status_log` Table**: Referenced to retrieve the full history of an order's status changes. (See [Order Service Migrations](#migrations) for definition).
- **`workers` Table**: Referenced to get the status of all registered kitchen workers. (See [Kitchen Worker Migrations](#migrations-1) for definition).

### API

#### Retrieves the current status and details of a single order.

```http
GET /orders/{order_number}/status
```

#### Response Body (`application/json`)

```json
{
  "order_number": "ORD_20241216_001",
  "current_status": "cooking",
  "updated_at": "2024-12-16T10:32:00Z",
  "estimated_completion": "2024-12-16T10:42:00Z",
  "processed_by": "chef_mario"
}
```

#### Retrieves the complete lifecycle history of an order.

```http
GET /orders/{order_number}/history
```

#### Response Body (`application/json`)

```json
[
  {
    "status": "received",
    "timestamp": "2024-12-16T10:30:00Z",
    "changed_by": "order-service"
  },
  {
    "status": "cooking",
    "timestamp": "2024-12-16T10:32:00Z",
    "changed_by": "chef_mario"
  }
]
```

#### Retrieves the status of all registered kitchen workers.

```http
GET /workers/status
```

#### Response Body (`application/json`)

```json
[
  {
    "worker_name": "chef_mario",
    "status": "online",
    "orders_processed": 5,
    "last_seen": "2024-12-16T10:35:00Z"
  },
  {
    "worker_name": "chef_luigi",
    "status": "offline",
    "orders_processed": 3,
    "last_seen": "2024-12-16T10:30:01Z"
  }
]
```

### Logic

1. **Receive Request:** The service listens for HTTP `GET` requests on its defined endpoints.

2. **Query Database:**

- For `/orders/{order_number}/status`, it queries the `orders` table for the specified `order_number`.
- For `/orders/{order_number}/history`, it queries the `order_status_log` table, joining with `orders` to find the correct order.
- For `/workers/status`, it queries the `workers` table. It should also include logic to determine if a worker is `offline` by checking if `now() - last_seen` exceeds a certain threshold (e.g., `2 * heartbeat-interval`).

3. **Format Response:** It formats the query results into the specified JSON structure.

4. **Error Handling:** If an `order_number` is not found, it should return an HTTP `404 Not Found`. For database errors, it should return an HTTP `500 Internal Server Error`.

### Logs

| Level | Action           | Description                                                    |
| ----- | ---------------- | -------------------------------------------------------------- |
| INFO  | service_started  | On successful startup.                                         |
| DEBUG | request_received | On receiving any API request. Include endpoint and parameters. |
| ERROR | db_query_failed  | If any database query fails.                                   |

### Flags

| Option   | Type   | Default | Description                                      |
| -------- | ------ | ------- | ------------------------------------------------ |
| `--mode` | string | —       | **Required.** Must be set to `tracking-service`. |
| `--port` | int    | `3002`  | The HTTP port for the API.                       |

### Usage

1. **Launch the service:**

   ```sh
   ./restaurant-system --mode=tracking-service --port=3002
   ```

2. **Query for an order's status:**

   ```sh
   curl http://localhost:3002/orders/ORD_20241216_001/status
   ```

3. **Query for worker status:**

   ```sh
   curl http://localhost:3002/workers/status
   ```

---

## Feature: Notification Service

### Context

The Notification Service is a simple subscriber that demonstrates the fanout capabilities of the messaging system. It listens for all order status updates published by the Kitchen Workers and displays them. In a real-world scenario, this service could be extended to send push notifications, emails, or SMS messages to customers.

### Migrations

This service does not interact with the database.

### API

This service does not expose any API endpoints. It outputs information to the console.

### Logic

1. **Startup and Connection:**
    - Connects to RabbitMQ.
    - Ensures the `notifications_fanout` exchange and the `notifications_queue` exist and that the queue is bound to the exchange.
2. **Message Consumption:**
    - Consumes `Status Update Messages` from the `notifications_queue`. The message format is defined in the [Kitchen Worker Logic](#logic-1) section.
3. **Display Notification:**
    - Upon receiving a message, it parses the JSON and prints a human-readable notification to standard output.
    - Example: `Notification for order ORD_20241216_001: Status changed from 'received' to 'cooking' by chef_mario.`
4. **Message Acknowledgement:** It sends a `basic.ack` to RabbitMQ to confirm the message has been processed and can be deleted from the queue.

### Logs

| Level | Action                     | Description                                                                                         |
| ----- | -------------------------- | --------------------------------------------------------------------------------------------------- |
| INFO  | service_started            | On successful startup.                                                                              |
| DEBUG | notification_received      | When a status update message is received. Include key details like `order_number` and `new_status`. |
| ERROR | rabbitmq_connection_failed | If the connection to RabbitMQ fails.                                                                |

### Flags

| Option   | Type   | Default | Description                                             |
| -------- | ------ | ------- | ------------------------------------------------------- |
| `--mode` | string | —       | **Required.** Must be set to `notification-subscriber`. |

### Usage

1. **Launch one or more subscribers:**

   ```sh
   # Terminal 1
   ./restaurant-system --mode=notification-subscriber

   # Terminal 2
   ./restaurant-system --mode=notification-subscriber
   ```

2. **Observe the output:** As Kitchen Workers process orders, both running subscribers will print the same notification messages to their respective consoles, demonstrating the broadcast pattern.
   _Expected Log Output:_
   ```json
   {
     "timestamp": "2024-12-16T10:32:05.000Z",
     "level": "INFO",
     "service": "notification-subscriber",
     "hostname": "notification-sub-1",
     "request_id": "a1b2c3d4e5f6",
     "action": "notification_received",
     "message": "Received status update for order ORD_20241216_001",
     "details": { "order_number": "ORD_20241216_001", "new_status": "cooking" }
   }
   ```
   _Expected Console Output:_
   `Notification for order ORD_20241216_001: Status changed from 'received' to 'cooking' by chef_mario.`

---