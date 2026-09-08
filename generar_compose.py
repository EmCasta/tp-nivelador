import sys
import os

"""
Opcional: Definir un script simple en bash, Python o Golang que 
reciba como parámetro la cantidad de clientes a configurar y 
genere automáticamente un nuevo archivo docker-compose.yaml.
"""

FILENAME = "docker-compose.yaml"

class Config:
    def __init__(self):
        self.server_host="server"
        self.server_port="5678"
        self.agency_quorum_min=3
        self.batch_size=10
        self.input_file="/input/input-0.csv"

    def obtener_entorno(self):
        self.server_host=os.getenv("SERVER_HOST", self.server_host)
        self.server_port=int(os.getenv("SERVER_PORT", self.server_port))
        self.agency_quorum_min=int(os.getenv("AGENCY_QUORUM_MIN", self.agency_quorum_min))
        self.batch_size=int(os.getenv("BATCH_SIZE", self.batch_size))
        self.input_file=os.getenv("INPUT_FILE", self.input_file)

def generar_server(config: Config):
    server = f"""services:
 server:
  build:
   context: ./services/server
   dockerfile: Dockerfile
  container_name: {config.server_host}
  environment:
   - PYTHONUNBUFFERED=1
   - SERVER_HOST={config.server_host}
   - SERVER_PORT={config.server_port}
   - AGENCY_QUORUM_MIN={config.agency_quorum_min}
  ports:
   - \"{config.server_port}:{config.server_port}\"

"""
    with open(FILENAME, "w") as archivo:
        archivo.write(server)

def generar_cliente(numero: int, config: Config):
    cliente = f""" client_{numero}:
  build:
   context: ./services/client
   dockerfile: Dockerfile
  container_name: client_{numero}
  depends_on:
   - server
  environment:
   - AGENCY_ID={numero}
   - SERVER_HOST={config.server_host}
   - SERVER_PORT={config.server_port}
   - INPUT_FILE={config.input_file}
   - OUTPUT_FILE=/output/output-{numero}.out
   - BATCH_SIZE={config.batch_size}
  volumes:
   - ./input:/input
   - ./output:/output

"""
    with open(FILENAME, "a") as archivo:
        archivo.write(cliente)

def main():
    config = Config()

    if len(sys.argv) != 2:
        print("Cantidad de argumentos invalida", file=sys.stderr)
        return

    cantidad_clientes = sys.argv[1]
    if not cantidad_clientes.isdigit():
        print("Cantidad de clientes a configurar debe ser numerica", file=sys.stderr)
        return
    cantidad_clientes = int(cantidad_clientes)
    if cantidad_clientes <= 0:
        print("Cantidad de clientes a configurar debe ser >= 0", file=sys.stderr)
        return

    try:
        config.obtener_entorno()
    except Exception:
        print("Ocurrio un error al setear variables de entorno", file=sys.stderr)
        return

    generar_server(config)
    for i in range(cantidad_clientes):
        generar_cliente(i, config)
    print(f"{FILENAME} creado exitosamente")

if __name__ == "__main__":
    main()
