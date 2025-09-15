# Imagens
SERVER_IMAGE = go_server
CLIENT_IMAGE = go_client
STRESS_IMAGE = go_stress

# Containers
SERVER_CONTAINER = server_container
CLIENT_CONTAINER = client_container
STRESS_CONTAINER = stress_container

# Dockerfiles
SERVER_DOCKERFILE = Dockerfile.server
CLIENT_DOCKERFILE = Dockerfile.client
STRESS_DOCKERFILE = Dockerfile.stress

# ================================
# Build
# ================================
build: build-server build-client build-stress

build-server:
	docker build -t $(SERVER_IMAGE) -f $(SERVER_DOCKERFILE) .

build-client:
	docker build -t $(CLIENT_IMAGE) -f $(CLIENT_DOCKERFILE) .

build-stress:
	docker build -t $(STRESS_IMAGE) -f $(STRESS_DOCKERFILE) .

# ================================
# Run
# ================================
run-server:
	docker run --name $(SERVER_CONTAINER) -p 8080:8080 $(SERVER_IMAGE)

run-client:
	docker run --name $(CLIENT_CONTAINER) $(CLIENT_IMAGE)

run-stress:
	docker run --name $(STRESS_CONTAINER) $(STRESS_IMAGE)

# ================================
# Cleanup
# ================================
stop:
	-docker stop $(SERVER_CONTAINER) $(CLIENT_CONTAINER) $(STRESS_CONTAINER)

rm:
	-docker rm -f $(SERVER_CONTAINER) $(CLIENT_CONTAINER) $(STRESS_CONTAINER)

clean: stop rm
	-docker rmi -f $(SERVER_IMAGE) $(CLIENT_IMAGE) $(STRESS_IMAGE)

# Conveniência
server: build-server run-server
client: build-client run-client
stress: build-stress run-stress
