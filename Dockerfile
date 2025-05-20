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

WORKDIR /src/c

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

RUN git clone --recurse-submodules https://github.com/mlc-ai/tokenizers-cpp.git
RUN . /root/.cargo/env && cd tokenizers-cpp && mkdir build && cd build && cmake .. && make -j$(nproc)

ARG RKLLM_VERSION=release-v1.2.1b1
ARG RKLLM_WRAPPER_VERSION=v1.2.1b1-1

RUN wget https://raw.githubusercontent.com/airockchip/rknn-llm/refs/tags/${RKLLM_VERSION}/rkllm-runtime/Linux/librkllm_api/include/rkllm.h -O /usr/include/rkllm.h
RUN wget https://github.com/airockchip/rknn-llm/raw/refs/tags/${RKLLM_VERSION}/rkllm-runtime/Linux/librkllm_api/aarch64/librkllmrt.so -O /usr/lib/librkllmrt.so


RUN git clone https://github.com/Tech-Arch1tect/rkllmwrapper-go.git -b ${RKLLM_WRAPPER_VERSION}
RUN cd rkllmwrapper-go/wrapper && bash build.sh


FROM golang:1.24-bookworm AS go-builder
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY --from=c-builder /src/c/rkllmwrapper-go/wrapper/librkllm_wrapper.so /usr/lib/
COPY --from=c-builder /src/c/rkllmwrapper-go/wrapper/rkllm_wrapper.h /usr/include/
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

COPY --from=c-builder /src/c/rkllmwrapper-go/wrapper/librkllm_wrapper.so /usr/lib/
COPY --from=c-builder /src/c/rkllmwrapper-go/wrapper/rkllm_wrapper.h /usr/include/
COPY --from=c-builder /usr/lib/librkllmrt.so /usr/lib/
COPY --from=c-builder /usr/include/rkllm.h /usr/include/
COPY --from=c-builder /src/c/tokenizers-cpp/build/libtokenizers_c.a /usr/lib/

WORKDIR /app
COPY --from=go-builder /go/src/app/bin/app .

EXPOSE 8080
ENTRYPOINT ["./app"]
