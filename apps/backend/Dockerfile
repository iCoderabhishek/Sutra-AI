FROM oven/bun:1

RUN apt-get update && \
    apt-get install -y docker.io git && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY package.json bun.lock* turbo.json ./
COPY packages/db/package.json packages/db/
COPY apps/backend/package.json apps/backend/

RUN bun install

COPY packages/db packages/db/
COPY apps/backend apps/backend/

RUN cd packages/db && bun db:generate

WORKDIR /app/apps/backend

ENV PORT=8000
EXPOSE $PORT

CMD sh -c "bun run src/api/index.ts"
