FROM debian:bookworm-slim AS c-builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      build-essential \
      ca-certificates \
      cmake \
      git \
      wget \
      libc6-dev       \
      libstdc++-12-dev      \
      curl \
      && rm -rf /var/lib/apt/lists/*

RUN wget https://raw.githubusercontent.com/airockchip/rknn-llm/refs/tags/release-v1.2.0/rkllm-runtime/Linux/librkllm_api/include/rkllm.h -O /usr/include/rkllm.h
RUN wget https://github.com/airockchip/rknn-llm/raw/refs/tags/release-v1.2.0/rkllm-runtime/Linux/librkllm_api/aarch64/librkllmrt.so -O /usr/lib/librkllmrt.so

WORKDIR /src/c
RUN git clone https://github.com/Tech-Arch1tect/rkllmwrapper.git
RUN cd rkllmwrapper && bash build.sh

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

RUN git clone --recurse-submodules https://github.com/mlc-ai/tokenizers-cpp.git
RUN . /root/.cargo/env && cd tokenizers-cpp && mkdir build && cd build && cmake .. && make -j$(nproc)

FROM golang:1.24-bookworm AS go-builder
WORKDIR /go/src/app
COPY go.mod go.sum .
RUN go mod download

COPY --from=c-builder /src/c/rkllmwrapper/librkllm_wrapper.so /usr/lib/
COPY --from=c-builder /src/c/rkllmwrapper/rkllm_wrapper.h /usr/include/
COPY --from=c-builder /usr/lib/librkllmrt.so /usr/lib/
COPY --from=c-builder /usr/include/rkllm.h /usr/include/   
COPY --from=c-builder /src/c/tokenizers-cpp/build/libtokenizers_c.a /usr/lib/
COPY --from=c-builder /src/c/tokenizers-cpp/include/tokenizers_c.h /usr/include/

COPY . .

RUN CGO_ENABLED=1 go build -o bin/app .

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      libgomp1 \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=c-builder /src/c/rkllmwrapper/librkllm_wrapper.so /usr/lib/
COPY --from=c-builder /src/c/rkllmwrapper/rkllm_wrapper.h /usr/include/
COPY --from=c-builder /usr/lib/librkllmrt.so /usr/lib/
COPY --from=c-builder /usr/include/rkllm.h /usr/include/
COPY --from=c-builder /src/c/tokenizers-cpp/build/libtokenizers_c.a /usr/lib/

WORKDIR /app
COPY --from=go-builder /go/src/app/bin/app .

EXPOSE 8080
ENTRYPOINT ["./app"]
