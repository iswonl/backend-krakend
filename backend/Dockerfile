# Build Stage
FROM rustlang/rust:nightly-slim AS builder
WORKDIR /home/rust/src

WORKDIR /home/rust/src/backend
COPY ./src ./src
COPY ./Cargo.toml ./

RUN apt-get update
RUN cargo install --path .

FROM ubuntu:latest
COPY --from=builder /usr/local/cargo/bin/backend ./backend
EXPOSE 8000

CMD ["./backend"]