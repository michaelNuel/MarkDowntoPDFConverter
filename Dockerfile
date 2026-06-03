# --- STAGE 1: Compile the Go binary ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency manifests and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code and compile the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o md2pdf main.go

# --- STAGE 2: Create the minimal runner image ---
FROM alpine:latest

# Install Chromium and standard text fonts
# (If we don't install ttf-freefont, Chrome won't render text inside the PDF!)
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont

WORKDIR /app

# Copy the compiled binary from Stage 1
COPY --from=builder /app/md2pdf .

# Expose port 8080 (standard HTTP port)
EXPOSE 8080

# In production web hosting, we want to run the server. 
# Our serve command requires an input file to watch. So we create a blank
# dummy.md file in the container to satisfy the program validation.
RUN touch dummy.md

# Start the server command on port 8080
CMD ["./md2pdf", "serve", "-in", "dummy.md", "-port", "8080"]