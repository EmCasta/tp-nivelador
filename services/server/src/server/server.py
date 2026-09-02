import socket
import logger
import safe_socket
from protocol.hello_packet import hello_packet_from_bytes
from protocol.bet_info_packet import bet_info_from_bytes, BetInfoPacket
from protocol.end_packet import end_from_bytes, EndPacket
from protocol.packet import TYPE_BET, TYPE_END
from lottery import Lottery
import traceback

STORAGE_PATH = "storage.tmp"

# TODO: manejar errores
# TODO: ver tema storage

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port

    def _handle_client(self, client_socket):
        action = "handle-client"
        try:
            logger.info(action, logger.LogResult.in_progress)
            lottery = Lottery(STORAGE_PATH)
            # recibir hello packet
            message_length = safe_socket.recv_all(
                client_socket, 1
            )
            length = int.from_bytes(message_length, "big", signed=False)
            packet = safe_socket.recv_all(
                client_socket, length
            )
            hello_packet = hello_packet_from_bytes(packet)
            client_agency_id = hello_packet.agency_id
            logger.info("hello-packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id, "actual-len", len(packet))

            while True:
                # recibir tamaño primero
                message_length = safe_socket.recv_all(
                    client_socket, 1
                )
                length = int.from_bytes(message_length, "big", signed=False)
                logger.info("msg-len-received", logger.LogResult.in_progress, "agency-id", client_agency_id, "len", length)

                # recibir paquete
                packet = safe_socket.recv_all(
                    client_socket, length
                )
                logger.info("packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id, "type", packet[0], "actual-len", len(packet))

                if packet[0] == TYPE_BET:
                    bet_info = bet_info_from_bytes(packet, client_agency_id)
                    logger.info("bet-packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id)

                    bet = bet_info.bet
                    lottery.store_bets([bet])
                if packet[0] ==  TYPE_END:
                    logger.info("end-packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id)

                    _ = end_from_bytes(packet)
                    # cliente termino de mandar las bets, calcular ganadores y mandarlos
                    break

            logger.info("sending-winners", logger.LogResult.in_progress, "agency-id", client_agency_id)
            for bet in lottery.load_bets():
                if lottery.has_won(bet) and bet.agency_id == client_agency_id:
                    # enviar bet a cliente!
                    bet_info = BetInfoPacket(bet)
                    safe_socket.send_all(client_socket, bet_info.serialize())
            # enviar end y terminar conexion
            end_packet = EndPacket()
            safe_socket.send_all(client_socket, end_packet.serialize())
            logger.info("end-packet-sent", logger.LogResult.in_progress, "agency-id", client_agency_id)

                
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail,
                "err", e, "traceback", traceback.format_exc())
            raise e

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
