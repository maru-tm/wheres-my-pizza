[] Добавить еррор хендлинг для воркеров ( чтобы дайн ин не принимал заказы тейкаут и не возникал) (
{"timestamp":"2025-09-15T12:29:10Z","level":"ERROR","service":"kitchen-worker","action":"order_processing_failed","message":"failed to process order","hostname":"Aida","request_id"
:"startup-19884-1757939233389037500","details":{"order_number":"ORD_20250915_004"},"error":{"msg":"worker cannot handle order type takeout","stack":"goroutine 1 [running]:\nruntime
/debug.Stack()\n\tC:/Program Files/Go/src/runtime/debug/stack.go:26 +0x5e\nrestaurant-system/pkg/logger.Log({0x7ff7ab514995, 0x5}, {0x7ff7ab518bec, 0xe}, {0x7ff7ab51e42e, 0x17}, {0
x7ff7ab51e445, 0x17}, {0xc000152300, 0x21}, ...)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/pkg/logger/logger.go:59 +0x229\nrestaurant-system/cmd/kitchen.Run({0x7ff7ab
5d5b08, 0xc00017c1e0}, 0xc000115790, 0xc00017c3c0, {0xc000126056, 0x5}, {0xc00010eb40, 0x1, 0x1}, 0x1, ...)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/cmd/kitchen/run.go:107 +0xbaa\nmain.main()\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/main.go:91 +0xc95\n"}}
)
[] Добавить валидацию данных (чтобы при не указании некоторых атрибутов не выводить прям такой длинный текст а просто выводить ерррор джисон и все) (
{"timestamp":"2025-09-15T12:47:58Z","level":"ERROR","service":"order-service","action":"validation_failed","message":"order validation failed","hostname":"Aida","request_id":"req-1
757940478621277600","details":{"error":"validation error: delivery_address is required for delivery orders"},"error":{"msg":"validation error: delivery_address is required for deli
very orders","stack":"goroutine 90 [running]:\nruntime/debug.Stack()\n\tC:/Program Files/Go/src/runtime/debug/stack.go:26 +0x5e\nrestaurant-system/pkg/logger.Log({0x7ff7ab514995, 0
x5}, {0x7ff7ab5183a2, 0xd}, {0x7ff7ab51aa08, 0x11}, {0x7ff7ab51e8f1, 0x17}, {0xc0000b4768, 0x17}, ...)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/pkg/logger/logger.go:
59 +0x229\nrestaurant-system/internal/order/service.(*OrderService).CreateOrder(0xc000320340, {0x7ff7ab5d5ad0, 0xc00040c210}, 0xc0000d83c0)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/
wheres-my-pizza/internal/order/service/order.go:47 +0x405\nrestaurant-system/internal/order/handler.(*OrderHandler).CreateOrderHandler(0xc00031a068, {0x7ff7ab5d4b00, 0xc00040a0f0},
0xc00024b998)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/internal/order/handler/handler.go:49 +0x51a\nrestaurant-system/cmd/order.Run.func1({0x7ff7ab5d4b00, 0xc00040a
0f0}, 0xc0002ae280)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/cmd/order/run.go:40 +0x185\nnet/http.HandlerFunc.ServeHTTP(0xc00036c0c0?, {0x7ff7ab5d4b00?, 0xc00040a0f0
?}, 0xc00024bb58?)\n\tC:/Program Files/Go/src/net/http/server.go:2322 +0x29\nnet/http.(*ServeMux).ServeHTTP(0x7ff7aaf66885?, {0x7ff7ab5d4b00, 0xc00040a0f0}, 0xc0002ae280)\n\tC:/Pro
gram Files/Go/src/net/http/server.go:2861 +0x1c7\nnet/http.serverHandler.ServeHTTP({0x7ff7ab5d24e0?}, {0x7ff7ab5d4b00?, 0xc00040a0f0?}, 0x6?)\n\tC:/Program Files/Go/src/net/http/se
rver.go:3340 +0x8e\nnet/http.(*conn).serve(0xc00023c360, {0x7ff7ab5d5ad0, 0xc0003191d0})\n\tC:/Program Files/Go/src/net/http/server.go:2109 +0x665\ncreated by net/http.(*Server).Serve in goroutine 78\n\tC:/Program Files/Go/src/net/http/server.go:3493 +0x485\n"}}
{"timestamp":"2025-09-15T12:47:58Z","level":"ERROR","service":"order-service","action":"order_creation_failed","message":"failed to create order","hostname":"Aida","request_id":"re
q-1757940478621277600","details":{"customer_name":"Jane Doe"},"error":{"msg":"validation error","stack":"goroutine 90 [running]:\nruntime/debug.Stack()\n\tC:/Program Files/Go/src/r
untime/debug/stack.go:26 +0x5e\nrestaurant-system/pkg/logger.Log({0x7ff7ab514995, 0x5}, {0x7ff7ab5183a2, 0xd}, {0x7ff7ab51d268, 0x15}, {0x7ff7ab51dcd3, 0x16}, {0xc0000b4768, 0x17},
...)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/pkg/logger/logger.go:59 +0x229\nrestaurant-system/internal/order/handler.(*OrderHandler).CreateOrderHandler(0xc00031a0
68, {0x7ff7ab5d4b00, 0xc00040a0f0}, 0xc00024b998)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/internal/order/handler/handler.go:51 +0xa0a\nrestaurant-system/cmd/order.R
un.func1({0x7ff7ab5d4b00, 0xc00040a0f0}, 0xc0002ae280)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/cmd/order/run.go:40 +0x185\nnet/http.HandlerFunc.ServeHTTP(0xc00036c0
c0?, {0x7ff7ab5d4b00?, 0xc00040a0f0?}, 0xc00024bb58?)\n\tC:/Program Files/Go/src/net/http/server.go:2322 +0x29\nnet/http.(*ServeMux).ServeHTTP(0x7ff7aaf66885?, {0x7ff7ab5d4b00, 0xc
00040a0f0}, 0xc0002ae280)\n\tC:/Program Files/Go/src/net/http/server.go:2861 +0x1c7\nnet/http.serverHandler.ServeHTTP({0x7ff7ab5d24e0?}, {0x7ff7ab5d4b00?, 0xc00040a0f0?}, 0x6?)\n\t
C:/Program Files/Go/src/net/http/server.go:3340 +0x8e\nnet/http.(*conn).serve(0xc00023c360, {0x7ff7ab5d5ad0, 0xc0003191d0})\n\tC:/Program Files/Go/src/net/http/server.go:2109 +0x665\ncreated by net/http.(*Server).Serve in goroutine 78\n\tC:/Program Files/Go/src/net/http/server.go:3493 +0x485\n"}}
)