from abc import ABC, abstractmethod

TYPE_HELLO = 0x00
TYPE_BET = 0x01
TYPE_END = 0x02

class Packet(ABC):

    @abstractmethod
    def get_type(self) -> int:
        pass

    @abstractmethod
    def serialize(self) -> bytes:
        pass

    def header(self) -> bytes:
        return self.get_type().to_bytes(1, "big", signed=False)
