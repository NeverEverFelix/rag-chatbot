FROM bitnami/postgresql:15

# Install git and build deps
USER root

RUN install_packages git make gcc libc-dev postgresql-server-dev-15

# Clone and build pgvector
RUN git clone --branch v0.5.1 https://github.com/pgvector/pgvector.git \
  && cd pgvector \
  && make \
  && make install \
  && cd .. && rm -rf pgvector

# Copy SQL init script to auto-enable pgvector at container startup
COPY pgvector-init.sql /docker-entrypoint-initdb.d/

# Revert to non-root user
USER 1001
