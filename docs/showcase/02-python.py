from dataclasses import dataclass

@dataclass(frozen=True)
class Shot:
    theme: str = "nightOwl"
    scale: int = 2

    def render(self, code: str) -> bytes:
        card = layout(highlight(code, self.theme))
        return rasterize(card, scale=self.scale)
