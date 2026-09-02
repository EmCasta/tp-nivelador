TYPE_HELLO = 0x00
TYPE_BET = 0x01
TYPE_END = 0x02

class Packet():

    def get_type(self) -> int:
        raise NotImplementedError("Subclase debe implementar get_type")

    def serialize(self) -> bytes:
        raise NotImplementedError("Subclase debe implementar serialize")

    def header(self) -> bytes:
        return self.get_type().to_bytes(1, "big", signed=False)
