infra-up:
	docker compose \
		-f infra/base.yaml \
		up -d --remove-orphans

infra-down:
	docker compose \
		-f infra/base.yaml \
		down --remove-orphans
		