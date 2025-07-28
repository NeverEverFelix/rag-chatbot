import os
import requests
import psycopg2
import textwrap

# Constants
EMBEDDING_URL = os.getenv("EMBEDDING_SERVICE_URL", "http://localhost:5001/embed")
POSTGRES_DSN = os.getenv("POSTGRES_DSN", "postgres://postgres:jCfAaeEOZJ@localhost:5432/ragdb")

CHUNK_SIZE = 500  # characters
OVERLAP = 50


def load_case_study(filepath: str) -> str:
    with open(filepath, "r", encoding="utf-8") as f:
        return f.read()

def chunk_text(text: str, chunk_size=CHUNK_SIZE, overlap=OVERLAP) -> list:
    chunks = []
    start = 0
    while start < len(text):
        end = start + chunk_size
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        start += chunk_size - overlap
    return chunks
def get_embedding(text: str) -> list:
    payload = {"text": text}
    try:
        response = requests.post(EMBEDDING_URL, json=payload)
        response.raise_for_status()
        return response.json()["embedding"]
    except Exception as e:
        print(f"Embedding failed for chunk: {text[:50]}... — {e}")
        return None
def insert_embedding(conn, chunk: str, embedding: list):
    vector_str = "[" + ",".join(f"{x:.6f}" for x in embedding) + "]"
    with conn.cursor() as cur:
        cur.execute(
            "INSERT INTO embeddings (chunk, embedding) VALUES (%s, %s::vector)",
            (chunk, vector_str)
        )
def embed_and_store_all(filepath: str):
    raw_text = load_case_study(filepath)
    chunks = chunk_text(raw_text)
    print(f"🧠 Chunked into {len(chunks)} segments.")

    conn = psycopg2.connect(POSTGRES_DSN)
    for i, chunk in enumerate(chunks):
        embedding = get_embedding(chunk)
        if embedding:
            insert_embedding(conn, chunk, embedding)
            print(f"✅ Inserted chunk {i+1}/{len(chunks)}")
    conn.commit()
    conn.close()
    print("🚀 All chunks embedded and stored.")
if __name__ == "__main__":
    embed_and_store_all("ci_cd_case_study.txt")  # or .md if that's what you saved
