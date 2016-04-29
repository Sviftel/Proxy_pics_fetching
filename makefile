SRCS = proxy_server.go \
       templates.go \
       errors.go \
       html_processing.go \
       network_processing.go

all: $(SRCS)
	go build $(SRCS)

clean:
	rm proxy_server
