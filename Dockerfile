FROM node:24-bookworm AS frontend
WORKDIR /build
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY vite.config.ts tsconfig.json ./
COPY web ./web
RUN pnpm exec vp build

FROM golang:1.27.0-bookworm AS plugin
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY --from=frontend /build/internal/console/dist ./internal/console/dist
RUN CGO_ENABLED=1 go build -trimpath -buildmode=c-shared -o /out/cpa2pool.so ./cmd/plugin

FROM scratch AS artifact
COPY --from=plugin /out/cpa2pool.so /cpa2pool.so
