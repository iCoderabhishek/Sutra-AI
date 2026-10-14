FROM oven/bun:1

# Prisma's query engine needs OpenSSL.
RUN apt-get update && \
    apt-get install -y --no-install-recommends openssl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY package.json bun.lock* turbo.json ./
COPY packages/db/package.json packages/db/
COPY apps/backend/package.json apps/backend/

RUN bun install

COPY packages/db packages/db/
COPY apps/backend apps/backend/

# prisma.config.ts requires DATABASE_URL, but generate never connects, so a
# placeholder is enough. The real URL comes from compose at runtime.
RUN cd packages/db && DATABASE_URL="postgresql://build:build@localhost:5432/build" bun db:generate

WORKDIR /app/apps/backend

ENV PORT=8000
EXPOSE $PORT

CMD sh -c "bun run src/api/index.ts"
