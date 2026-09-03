import socket
import threading
import signal
import logger
import safe_socket
from protocol.hello_packet import hello_packet_from_bytes
from protocol.bet_info_packet import bet_info_from_bytes, BetInfoPacket
from protocol.packet import TYPE_BET, LENGTH_BYTES, get_last_packet_flag, set_last_packet_flag
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
        signal.signal(signal.SIGTERM, self.shutdown_gracefully)

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            self.sockets.add(server_socket)
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                client_thread = threading.Thread(target=self._handle_client, args=(client_socket,))
                client_thread.start()
                self.threads.append(client_thread)

    def shutdown_gracefully(self):
        with self.sockets_lock:
            for sock in self.sockets:
                sock.close()
        for thread in self.threads:
            thread.join()
        # faltaria manejar el archivo!
    

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
        message_length = safe_socket.recv_all(
            client_socket, LENGTH_BYTES
        )
        length = int.from_bytes(message_length, "big", signed=False)
        packet = safe_socket.recv_all(
            client_socket, length
        )
        hello_packet = hello_packet_from_bytes(packet)
        return hello_packet

    def _receive_bets(self, client_socket, lottery, client_info):
        keep_receiving = True
        while keep_receiving:
            message_length = safe_socket.recv_all(
                client_socket, LENGTH_BYTES
            )
            length = int.from_bytes(message_length, "big", signed=False)
            packet = safe_socket.recv_all(
                client_socket, length
            )
            packet = bytearray(packet)

            is_last = get_last_packet_flag(packet, 0)
            keep_receiving = not is_last
            if packet[0] == TYPE_BET:
                bet_info = bet_info_from_bytes(packet, client_info.agency_id, client_info.batch_size)
                with self.file_lock:
                    lottery.store_bets(bet_info.bets)
            else:
                raise ValueError("Invalid packet type")

    def _send_winners(self, client_socket, lottery, client_info):
        self.quorum_barrier.wait()

        with self.file_lock:
            keep_sending = True
            bet_iterator = lottery.load_bets()
            while keep_sending:
                bets = []
                while len(bets) < client_info.batch_size:
                    bet = next(bet_iterator, None)
                    if bet is None:
                        keep_sending = False
                        break
                    if lottery.has_won(bet) and bet.agency_id == client_info.agency_id:
                        bets.append(bet)
                if len(bets) == 0:
                    return
                packet = BetInfoPacket(bets).serialize()
                if not keep_sending:
                    set_last_packet_flag(packet, LENGTH_BYTES)
                safe_socket.send_all(client_socket, packet)
