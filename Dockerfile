FROM golang:1.22-alpine

# If you do not have or want to point GO to your private repo to pull the needed packages, feel free to comment the below section(untill next comment).
ARG REPO_USER_TOKEN

RUN echo 'machine private.repo.io' >~/.netrc && \
    echo 'login repo-user' >>~/.netrc && \
    echo "password ${REPO_USER_TOKEN}" >>~/.netrc && \
    chmod 600 ~/.netrc

ENV GOSUMDB=off
ENV GOPROXY=https://private.repo.io/artifactory/api/go/golang-virtual,direct

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and source code into the container
COPY go.mod ./
COPY . .

# Install dependencies and build the application
RUN go mod tidy && \
    go build -o webhook-project .

# Expose port 8443
EXPOSE 8443

# Set the entrypoint to `go run .` to run the application on container startup
ENTRYPOINT ["go", "run", "."]
