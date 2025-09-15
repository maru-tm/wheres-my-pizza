[] Add error handling especially for workers (
{"timestamp":"2025-09-15T12:29:10Z","level":"ERROR","service":"kitchen-worker","action":"order_processing_failed","message":"failed to process order","hostname":"Aida","request_id"
:"startup-19884-1757939233389037500","details":{"order_number":"ORD_20250915_004"},"error":{"msg":"worker cannot handle order type takeout","stack":"goroutine 1 [running]:\nruntime
/debug.Stack()\n\tC:/Program Files/Go/src/runtime/debug/stack.go:26 +0x5e\nrestaurant-system/pkg/logger.Log({0x7ff7ab514995, 0x5}, {0x7ff7ab518bec, 0xe}, {0x7ff7ab51e42e, 0x17}, {0
x7ff7ab51e445, 0x17}, {0xc000152300, 0x21}, ...)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/pkg/logger/logger.go:59 +0x229\nrestaurant-system/cmd/kitchen.Run({0x7ff7ab
5d5b08, 0xc00017c1e0}, 0xc000115790, 0xc00017c3c0, {0xc000126056, 0x5}, {0xc00010eb40, 0x1, 0x1}, 0x1, ...)\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/cmd/kitchen/run.go:107 +0xbaa\nmain.main()\n\tC:/Users/Админ/OneDrive/Desktop/Alem/wheres-my-pizza/main.go:91 +0xc95\n"}}
)
[]