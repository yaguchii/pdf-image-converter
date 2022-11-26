FROM golang:1.19

# Ignore APT warnings about not having a TTY
ENV DEBIAN_FRONTEND noninteractive

# install build essentials
RUN apt-get update && \
    apt-get install -y wget build-essential pkg-config --no-install-recommends

# Install ImageMagick deps
RUN apt-get -q -y install libjpeg-dev libpng-dev libwebp-dev libtiff-dev \
    libgif-dev libx11-dev --no-install-recommends

# See: https://stackoverflow.com/questions/53569383/imagick-fails-to-execute-gs-command
RUN apt-get install -y ghostscript-x

# See: https://askubuntu.com/questions/251950/imagemagick-convert-cant-convert-to-webp
RUN apt-get install -y webp

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