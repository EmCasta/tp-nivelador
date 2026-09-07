class ClientInfo:
    """
    Encapsula la informacion recibida de un cliente
    """
    def __init__(self, agency_id, batch_size):
        self.agency_id = agency_id
        self.batch_size = batch_size
