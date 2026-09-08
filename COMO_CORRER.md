# Cómo correr

- Para generar un archivo *docker-compose.yaml* con un servidor y N clientes:

    ```bash
    ./generar_compose.py [cantidad-de-clientes]
    ```

    Además, el *script* permite configurar las variables de entorno:

    - `SERVER_HOST`
    - `SERVER_PORT`
    - `AGENCY_QUORUM_MIN`
    - `BATCH_SIZE`
    - `INPUT_FILE`

    Por ejemplo:

    ```bash
    BATCH_SIZE=5 ./generar_compose.py 3
    ```

    Se crea un `docker-compose.yaml` con 3 clientes, cada uno con un batch size de 5.

- Para correr los contenedores utilizando *docker-compose*:

    ```bash
    mkdir -p output
	rm ./output/* -f
    COMPOSE_HTTP_TIMEOUT=300 docker compose -f docker-compose.yaml up --build --remove-orphans --detach
    ```

    o también se puede hacer `make up`

- Para ver los logs:

    ```bash
    docker compose -f $(DOCKER_FILE_PATH) logs --follow
    ```

    o también se puede hacer `make logs`

- Para detener los contenedores:

    ```bash
    docker compose -f $(DOCKER_FILE_PATH) stop -t 5
	docker compose -f $(DOCKER_FILE_PATH) down
    ```

    o también se puede hacer `make down`

- Para correr los tests:

    ```bash
    rm failed_test.log -f
	PYTHONPATH="$(PWD)" python3 tests/run.py
    ```

    o también se puede hacer `make test`
