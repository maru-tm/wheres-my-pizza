build:
	go build -o restaurant-system .
docker-up:
	DOCKER_BUILDKIT=0 docker-compose build --no-cache
	docker-compose up -d
docker-down:
	docker-compose down -v
	docker volume prune -f
	docker network prune -f

order-service:
	./restaurant-system --mode=order-service --port=3000
kitchen-worker:
	./restaurant-system --mode=kitchen-worker --worker-name="Alex" --prefetch=""
	./restaurant-system --mode=kitchen-worker --worker-name="Jamie" --order-types="dine_in"
	./restaurant-system --mode=kitchen-worker --worker-name="Taylor" --order-types="delivery"
	./restaurant-system --mode=kitchen-worker --worker-name="Morgan" --order-types="takeout"
tracking-service:
	./restaurant-system --mode=tracking-service --port=3002
notification-subscriber:
	./restaurant-system --mode=notification-subscriber
