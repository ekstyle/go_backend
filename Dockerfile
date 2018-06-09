FROM golang
#Copy content
ADD . /go/src/go_backend

#Dependency
RUN go get -u github.com/gorilla/mux

# Run the outyet command by default when the container starts.
RUN go install go_backend
ENTRYPOINT /go/bin/go_backend

# Document that the service listens on port 8080.
EXPOSE 8080