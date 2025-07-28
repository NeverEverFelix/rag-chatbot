FROM python:3.11-slim

WORKDIR /app

COPY main.py .
COPY crud.py .
COPY database.py .

RUN pip install --no-cache-dir \
    fastapi \
    uvicorn \
    psycopg2-binary \
    requests \
    python-dotenv \
    sentence-transformers\
    sqlalchemy\
    pgvector


EXPOSE 5001
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "5001"]