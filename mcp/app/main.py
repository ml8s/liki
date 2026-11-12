"""LIKI-MCP gateway process."""

import os

import uvicorn
from app.gateway import create_app


def main() -> None:
    app = create_app()
    uvicorn.run(
        app,
        host=os.environ.get("LIKI_BIND_HOST", "127.0.0.1"),
        port=int(os.environ.get("LIKI_BIND_PORT", "8087")),
        lifespan="off",
    )


if __name__ == "__main__":
    main()
