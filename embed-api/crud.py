import os
from sqlalchemy import create_engine, Column, Integer, Text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.types import UserDefinedType
from pgvector.sqlalchemy import Vector

# --- Custom pgvector type ---
class Vector(UserDefinedType):
    def __init__(self, dims=384):
        self.dims = dims

    def get_col_spec(self):
        return f"vector({self.dims})"

# --- Enforce reading from Kubernetes-provided POSTGRES_DSN ---
DATABASE_URL = os.environ["POSTGRES_DSN"]
print(f"[embed-api] Connecting to DB at: {DATABASE_URL}")

# --- SQLAlchemy setup ---
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# --- Embedding model ---
class Embedding(Base):
    __tablename__ = "embeddings"

    id = Column(Integer, primary_key=True, index=True)
    chunk = Column(Text, nullable=False)
    embedding = Column(Vector(384), nullable=False)

# --- Create the table (only run once, or use Alembic later) ---
def create_tables():
    Base.metadata.create_all(bind=engine)

# --- Insert embedding ---
def insert_embedding(text: str, vector: list[float]):
    db = SessionLocal()
    try:
        embedding = Embedding(chunk=text, embedding=vector)
        db.add(embedding)
        db.commit()
        db.refresh(embedding)
        return embedding
    finally:
        db.close()
