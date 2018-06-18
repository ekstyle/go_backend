FROM golang
#Copy content
ADD . /go/src/github.com/ekstyle/go_backend

#Dependency
RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/gorilla/schema

# Run the outyet command by default when the container starts.
RUN go install github.com/ekstyle/go_backend
ENTRYPOINT /go/bin/go_backend

# Document that the service listens on port 8080.
EXPOSE 8080