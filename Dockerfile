# Install the transcoding tool. The node will be accessible via HTTP on port 8100 and the cli tool will be on the path. The port can be changed.

# This image will be created to use an entrypoint. If you need to create a container 
#   with a shell you can run:
#
#    docker run -i -t -entrypoint='/bin/bash' rethinkdb -i
#
#    Just keep in mind that's giving you a shell in a new instance of the image not 
#    connecting you to an already running container.
#
#
#    Otherwise the main use of the file is:
#    docker run transcoding \
#            320p \
#            -i sample.mp4 \
#            -o output.mp4

FROM jrottenberg/ffmpeg
MAINTAINER Farid Zakaria

# We now need to install Go & GCC
# Taken from:
# https://github.com/docker-library/golang/blob/1eab0db63794152b4516dbcb70270eb9dced4cbd/1.5/Dockerfile
# gcc for cgo
RUN yum -y update && yum -y install \
		g++ \
		gcc \
		libc6-dev \
		make \
		git

ENV GOLANG_VERSION 1.5.3
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 43afe0c5017e502630b1aea4d44b8a7f059bf60d7f29dfd58db454d4e4e0ae53

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Create a directory inside the container to store all our application and then make it the working directory.
RUN mkdir -p /go/src/github.com/fzakaria/transcoding
WORKDIR /go/src/github.com/fzakaria/transcoding

# Copy the application directory (where the Dockerfile lives) into the container.
COPY . /go/src/github.com/fzakaria/transcoding

# Fetch necessary dependencies
RUN go get github.com/fzakaria/transcoding/

# Install out application
RUN go install github.com/fzakaria/transcoding

# Set the PORT environment variable inside the container
ENV PORT 8080

# Expose port 8080 to the host so we can access our application
EXPOSE 8080


ENTRYPOINT ["transcoding"]
CMD ["--config", "/go/src/github.com/fzakaria/transcoding/configs/prod-us-east-1.toml"]