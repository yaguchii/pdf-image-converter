FROM golang:1.21

# install build essentials
RUN apt-get upgrade  \
    && apt-get update  \
    && apt-get install -y wget build-essential pkg-config --no-install-recommends \
    libjpeg-dev libpng-dev libwebp-dev libtiff-dev libgif-dev libx11-dev --no-install-recommends \
    ghostscript-x webp \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV IMAGEMAGICK_VERSION=7.0.8-11

RUN cd && \
	wget https://github.com/ImageMagick/ImageMagick/archive/${IMAGEMAGICK_VERSION}.tar.gz && \
	tar xvzf ${IMAGEMAGICK_VERSION}.tar.gz && \
	cd ImageMagick* && \
	./configure \
	    --without-magick-plus-plus \
	    --without-perl \
	    --disable-openmp \
	    --with-gvc=no \
	    --disable-docs && \
	make -j$(nproc) && make install && \
	ldconfig /usr/local/lib

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -o /main -v

EXPOSE 8080

ENTRYPOINT ["/main"]