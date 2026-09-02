import socket
import logger
import safe_socket
from protocol.hello_packet import hello_packet_from_bytes
from protocol.bet_info_packet import bet_info_from_bytes, BetInfoPacket
from protocol.packet import TYPE_BET, LENGTH_BYTES, get_last_packet_flag, set_last_packet_flag
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
                client_socket, LENGTH_BYTES
            )
            length = int.from_bytes(message_length, "big", signed=False)
            packet = safe_socket.recv_all(
                client_socket, length
            )
            hello_packet = hello_packet_from_bytes(packet)
            client_agency_id = hello_packet.agency_id
            batch_size = hello_packet.batch_size
            logger.info("hello-packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id, "batch_size", batch_size)

            keep_receiving = True
            while keep_receiving:
                # recibir tamaño primero
                message_length = safe_socket.recv_all(
                    client_socket, LENGTH_BYTES
                )
                length = int.from_bytes(message_length, "big", signed=False)
                logger.info("msg-len-received", logger.LogResult.in_progress, "agency-id", client_agency_id, "len", length)

                # recibir paquete
                packet = safe_socket.recv_all(
                    client_socket, length
                )
                packet = bytearray(packet)
                logger.info("packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id, "type", packet[0], "actual-len", len(packet))

                is_last = get_last_packet_flag(packet)
                keep_receiving = not is_last
                if packet[0] == TYPE_BET:
                    bet_info = bet_info_from_bytes(packet, client_agency_id, batch_size)
                    logger.info("bet-packet-received", logger.LogResult.in_progress, "agency-id", client_agency_id)
                    lottery.store_bets(bet_info.bets)
                else:
                    # TODO: revisar que hacer acá
                    return

            # enviar bets ganadoras
            logger.info("sending-winners", logger.LogResult.in_progress, "agency-id", client_agency_id)
            keep_sending = True
            bet_iterator = lottery.load_bets()
            while keep_sending:
                bets = []
                while len(bets) < batch_size:
                    bet = next(bet_iterator, None)
                    if bet is None:
                        keep_sending = False
                        break
                    if lottery.has_won(bet) and bet.agency_id == client_agency_id:
                        bets.append(bet)

                packet = BetInfoPacket(bets).serialize()
                if not keep_sending:
                    set_last_packet_flag(packet)
                safe_socket.send_all(client_socket, packet)

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
