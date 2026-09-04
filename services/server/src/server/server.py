import socket
import threading
import signal
import os
import logger
import safe_socket
from protocol.hello_packet import hello_packet_from_bytes
from protocol.bet_info_packet import bet_info_from_bytes, BetInfoPacket
from protocol.packet import TYPE_BET, LENGTH_BYTES, get_last_packet_flag, set_last_packet_flag
from protocol.ack_packet import AckPacket, TYPE_ACK
from server.utils import read_packet
from server.client_info import ClientInfo
from lottery import Lottery
import traceback

STORAGE_PATH = "storage.tmp"

# TODO: manejar errores
# TODO: ver tema storage

class Server:
    def __init__(self, server_host: str, server_port: int, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.agency_quorum_min = agency_quorum_min
        self.file_lock = threading.Lock()
        self.quorum_barrier = threading.Barrier(agency_quorum_min)
        self.threads = []
        self.sockets = set()
        self.sockets_lock = threading.Lock()
        self.is_running = True
        signal.signal(signal.SIGTERM, self.shutdown_gracefully)

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            self.sockets.add(server_socket)
            while self.is_running:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except OSError:
                    # si ocurre esta excepcion, el server cerro el socket y hay que terminar
                    logger.info(
                        action, logger.LogResult.in_progress,
                        "OSerror: closing server")
                    return
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                client_thread = threading.Thread(target=self._handle_client, args=(client_socket,))
                client_thread.start()
                self.threads.append(client_thread)

    def shutdown_gracefully(self, signum, frame):
        with self.sockets_lock:
            for sock in self.sockets:
                sock.close()
        self.quorum_barrier.abort()
        for thread in self.threads:
            thread.join()
        os.remove(STORAGE_PATH)
        self.is_running = False
        logger.info("shutdown", logger.LogResult.success)

    def _handle_client(self, client_socket):
        action = "handle-client"
        try:
            self._store_socket(client_socket)

            logger.info(action, logger.LogResult.in_progress)
            lottery = Lottery(STORAGE_PATH)

            hello_packet = self._wait_for_hello(client_socket)
            client_info = ClientInfo(hello_packet.agency_id, hello_packet.batch_size)
            logger.info("hello-packet-received", logger.LogResult.in_progress, "agency-id", client_info.agency_id, "batch_size", client_info.batch_size)

            self._receive_bets(client_socket, lottery, client_info)
            logger.info("bets-received", logger.LogResult.in_progress, "agency-id", client_info.agency_id)

            logger.info("sending-winners", logger.LogResult.in_progress, "agency-id", client_info.agency_id)
            self._send_winners(client_socket, lottery, client_info)

        except OSError:
            # si ocurre esta excepcion, el server cerro el socket y hay que terminar
            logger.info(
                action, logger.LogResult.in_progress,
                "OSerror: closing client")
            return
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail,
                "err", e, "traceback", traceback.format_exc())
            raise e

        finally:
            self._cleanup(client_socket)

    def _cleanup(self, client_socket):
        with self.sockets_lock:
            if client_socket in self.sockets:
                self.sockets.remove(client_socket)
        client_socket.close()

    def _store_socket(self, client_socket):
        with self.sockets_lock:
            self.sockets.add(client_socket)
        

    def _wait_for_hello(self, client_socket):
        # esperar mensaje de hello
        packet = read_packet(client_socket)
        hello_packet = hello_packet_from_bytes(packet)
        # enviar ack si llego correctamente
        ack = AckPacket().serialize()
        safe_socket.send_all(client_socket, ack)
        return hello_packet

    def _receive_bets(self, client_socket, lottery, client_info):
        keep_receiving = True
        while keep_receiving:
            # esperar packete con batch de bets
            packet = bytearray(read_packet(client_socket))

            is_last = get_last_packet_flag(packet, 0)
            keep_receiving = not is_last
            bet_info = bet_info_from_bytes(packet, client_info.agency_id, client_info.batch_size)
            with self.file_lock:
                lottery.store_bets(bet_info.bets)
            # todas las bets se procesaron correctamente: enviar ack
            ack = AckPacket().serialize()
            safe_socket.send_all(client_socket, ack)
            

    def _send_winners(self, client_socket, lottery, client_info):
        try:
            self.quorum_barrier.wait()
        except threading.BrokenBarrierError:
            return

        bets = []
        with self.file_lock:
            for bet in lottery.load_bets():
                if lottery.has_won(bet) and bet.agency_id == client_info.agency_id:
                    bets.append(bet)
        if len(bets) == 0:
            # no hay ganadores: mandar ack, esperar respuesta y terminar
            ack = AckPacket().serialize()
            safe_socket.send_all(client_socket, ack)
            packet = read_packet(client_socket)
            if packet[0] != TYPE_ACK:
                raise ValueError("Invalid packet type: ACK expected")
            return
        
        actual_offset = 0
        while actual_offset < len(bets):
            # enviar batches con ganadores, esperando ack para cada uno de ellos
            packet = BetInfoPacket(bets[actual_offset:actual_offset+client_info.batch_size]).serialize()
            actual_offset += client_info.batch_size
            if actual_offset >= len(bets):
                set_last_packet_flag(packet, LENGTH_BYTES)
            safe_socket.send_all(client_socket, packet)
            packet = read_packet(client_socket)
            if packet[0] != TYPE_ACK:
                raise ValueError("Invalid packet type: ACK expected")
