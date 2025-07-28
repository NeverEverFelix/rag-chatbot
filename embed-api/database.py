import os
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy import Column, Integer, String
from pgvector.sqlalchemy import Vector

# -------------------------------------------------------------------
# ✅ DATABASE URL — reads from env or defaults to K8s Postgres DNS
# -------------------------------------------------------------------
DATABASE_URL = os.getenv(
    "POSTGRES_DSN",  # matches your K8s env var
    "postgresql+asyncpg://NeverEverFelix:138824@pgvector-postgresql.vector-db.svc.cluster.local:5432/ragdb"
)

# -------------------------------------------------------------------
# ✅ Async engine and session
# -------------------------------------------------------------------
engine = create_async_engine(DATABASE_URL, echo=True, future=True)
AsyncSessionLocal = sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)
Base = declarative_base()

# -------------------------------------------------------------------
# ✅ Embedding model
# -------------------------------------------------------------------
class Embedding(Base):
    __tablename__ = "embeddings"

    id = Column(Integer, primary_key=True, index=True)
    chunk = Column(String, nullable=False)
    embedding = Column(Vector(1536))  # adjust if your model uses 768-dim
