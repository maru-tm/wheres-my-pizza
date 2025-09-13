docker-up:
	DOCKER_BUILDKIT=0 docker-compose build --no-cache
	docker-compose up
docker-down:
	docker-compose down -v
	docker volume prune -f
	docker network prune -f

order-service:
	./restaurant-system --mode=order-service --port=3000
kitchen-worker:
	./restaurant-system --mode=kitchen-worker --worker-name=chef_mario --order-types=dine_in,takeout
tracking-service:
	./restaurant-system --mode=tracking-service --port=3002
notification-subscriber:
	./restaurant-system --mode=notification-subscriber
